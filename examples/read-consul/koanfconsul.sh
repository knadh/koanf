consul agent -dev &

exitcode=1
iterations=0

while [ $exitcode -ne 0 ]
do
	consul members
	exitcode=$?

	sleep 1
	((iterations++))

	if [ $iterations -gt 5 ]; then
		break
	fi
done

if [ $exitcode -eq 0 ]; then
	echo -e "\nTest program running..."
	go run main.go
	echo -e "\nShutdown..."
	consul leave
else
	echo "Consul server is unavailable."
	consul leave
fi
