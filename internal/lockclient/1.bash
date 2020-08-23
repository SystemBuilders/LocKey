for i in {1..10}
do
	go test -race -run TestLockService >> logs
done
