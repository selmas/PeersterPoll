set -eu
set -o pipefail

export PATH="$PATH:$PWD:$PWD/client:$PWD/server"

wait_port_open() {
	local port=$1

	while ! netstat -l | grep :$port > /dev/null
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

	server -name "$name" -UIPort "$uiPort" -gossipAddr "127.0.0.1:$gossipPort" -peers "$peers" 2>&1 > "$name.log" &

	wait_port_open $uiPort
	wait_port_open $gossipPort
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

new_key() {
	local origin=$1

	client key new "$origin"
}

poll_new() {
	local port=$1
	shift

	client -UIPort $port poll new "$@"
}

poll_list_contains_id() {
	local port=$1
	local id=$2

	client -UIPort $port poll list | grep -q "$id"
}

vote_put() {
	local port=$1
	local id=$2
	local option=$3

	client -UIPort $port vote put "$id" "$option"
}

log_check() {
	local name=$1
	local pattern=$2

	grep "$pattern" "$name.log" > /dev/null
}

log_wait() {
	local name=$1
	local pattern=$2

	while ! log_check "$name" "$pattern"
	do
		sleep 0.2
	done
}

cleanup() {
	rm -rf *.log *.key keys

	pkill -x server
	wait 2>/dev/null || :
}
trap 'cleanup' EXIT

build_all
