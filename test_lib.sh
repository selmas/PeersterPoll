set -eu
set -o pipefail

export PATH="$PATH:$PWD:$PWD/client"

wait_port_open() {
	local port=$1

	while ! netstat -l | grep -q :$port
	do
		sleep 0.2
	done
}

start_server() {
	local name=$1
	local uiPort=$2
	local gossipPort=$3
	shift 3
	local peers="$(echo "$*" | tr ' ' _)"

	Peerster -name "$name" -UIPort "$uiPort" -gossipAddr "127.0.0.1:$gossipPort" -peers "$peers" 2>&1 > "$name.log" & > $name.log

	wait_port_open $uiPort
	# TODO buggy wait_port_open $gossipPort 
}

client_index_file() {
	local port=$1
	local filename=$2

	client -UIPort=$port -file="$filename"
}

client_get_file_from() {
	local port=$1
	local dest=$2
	local filename=$3
	local metahash=$4

	client -UIPort=$port -Dest="$dest" -file="$filename" -request="$metahash"
}

client_get_file() {
	local port=$1
	local filename=$2
	local metahash=$3

	client -UIPort=$port -file="$filename" -request="$metahash"
}

get_number_of_chunks() {
	local filename=$1

	readonly block_size=8192

	echo $(($(wc -c "$filename" | cut -d ' ' -f 1) / block_size))
}

get_empty_file() {
	local file=$(mktemp)

	ln -s "$file"
	echo "${file##*/}"
}

get_random_file() {
	local file=$(get_empty_file)

	dd if=/dev/urandom ibs=1K count=64 of="$file" 2> /dev/null

	echo "$file"
}

build_all() {
	local old_pwd=$PWD

	find -name '*.go' -exec dirname '{}' \; | sort -u | while read dir
	do
		cd "$dir"
		go build
		cd "$old_pwd"
	done
}

# TODO everything is saved to the same dir, so not that useful
check_downloaded() {
	local filename=$1

	cmp "$filename" "hw3/_Download/$filename"
}

log_check() {
	local name=$1
	local pattern=$2

	grep -q "$pattern" "$name.log"
}

search_file() {
	local port=$1
	shift
	local keywords="$(echo "$*" | tr ' ' ,)"

	client -UIPort=$port -keywords="$keywords"
}

cleanup() {
	rm -rf tmp.* chunk hw3 *.log

	pkill -x Peerster
	wait 2>/dev/null || :
}
trap 'cleanup' EXIT

build_all
