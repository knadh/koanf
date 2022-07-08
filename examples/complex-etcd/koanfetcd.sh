# Testing script. Requirements: installed etcd

# setting the directory for the etcd
CURRENT_DIR=`pwd`
ETCD_TESTDIR="$CURRENT_DIR/etcdtest"

etcd --data-dir "$ETCD_TESTDIR" &

etcdctl endpoint health --dial-timeout=4s

# main test
echo -e "\nChecking...\n"
go run main.go

ETCD_PID=`pidof etcd`
kill -9 $ETCD_PID

rm -rf "$ETCD_TESTDIR"
#rm k_etcd_check
