coverage:
	go tool cover -func=reports/auth_coverage.out
	go tool cover -func=reports/loc_coverage.out
	go tool cover -func=reports/port_coverage.out
	go tool cover -func=reports/vxc_coverage.out

login:
	go test ./authentication -run TestLogin

test-vxc:
	go test -v ./vxc -timeout 60m

test-partner-lookup:
	go test -v ./partner -timeout 60m

