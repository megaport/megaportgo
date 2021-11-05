TEST_TIMEOUT = -timeout 60m
INTEGRATION_TAG = integration
UNIT_TAG = unit

export LOG_LEVEL ?= INFO

#######################
# Unit Testing #
#######################

unit: clean-test-cache auth-unit vxc-unit

auth-unit:
	@echo "Unit Testing Authentication Package"
	go test ${TEST_TIMEOUT} -v ./service/authentication -tags ${UNIT_TAG}

vxc-unit:
	@echo "Unit Testing Authentication Package"
	go test ${TEST_TIMEOUT} -v ./service/vxc -tags ${UNIT_TAG}

#######################
# Integration Testing #
#######################

integration: clean-test-cache auth-integ location-integ mcr-integ partner-integ port-integ vxc-integ

auth-integ:
	@echo "Integration Testing Authentication Package; Log Level: ${LOG_LEVEL}"
	go test ${TEST_TIMEOUT} -v ./service/authentication -tags ${INTEGRATION_TAG}

location-integ:
	@echo "Integration Testing Location Package; Log Level: ${LOG_LEVEL}"
	go test ${TEST_TIMEOUT} -v ./service/location -tags ${INTEGRATION_TAG}

mcr-integ:
	@echo "Integration Testing MCR Package; Log Level: ${LOG_LEVEL}"
	go test ${TEST_TIMEOUT} -v ./service/mcr -tags ${INTEGRATION_TAG}

partner-integ:
	@echo "Integration Testing Partner Package; Log Level: ${LOG_LEVEL}"
	go test ${TEST_TIMEOUT} -v ./service/partner -tags ${INTEGRATION_TAG}

port-integ:
	@echo "Integration Testing Port Package; Log Level: ${LOG_LEVEL}"
	go test ${TEST_TIMEOUT} -v ./service/port -tags ${INTEGRATION_TAG}

vxc-integ:
	@echo "Integration Testing VXC Package; Log Level: ${LOG_LEVEL}"
	go test ${TEST_TIMEOUT} -v ./service/vxc -tags ${INTEGRATION_TAG}

#############
# Utilities #
#############

clean-test-cache:
	go clean -testcache

create-user:
	go run test/create-user.go

coverage:
	go tool cover -func=reports/auth_coverage.out
	go tool cover -func=reports/loc_coverage.out
	go tool cover -func=reports/port_coverage.out
	go tool cover -func=reports/vxc_coverage.out