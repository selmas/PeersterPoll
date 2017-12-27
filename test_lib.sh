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

	Peerster -name "$name" -UIPort "$uiPort" -gossipAddr "127.0.0.1:$gossipPort" -peers "$peers" 2>&1 > "$name.log" &

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

log_check() {
	local name=$1
	local pattern=$2

	grep -q "$pattern" "$name.log"
}

cleanup() {
	rm -rf tmp.* chunk hw3 *.log

	pkill -x Peerster
	wait 2>/dev/null || :
}
trap 'cleanup' EXIT

build_all
