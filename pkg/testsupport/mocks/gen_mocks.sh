rm -rf pkg/testsupport/mocks/mock_*.go

bin/mockgen -source pkg/services/redis_store.go -destination pkg/testsupport/mocks/mock_redis_store.go -package mocks
bin/mockgen -source pkg/services/time_service.go -destination pkg/testsupport/mocks/mock_time_service.go -package mocks
bin/mockgen -source pkg/services/inhooks_config_service.go -destination pkg/testsupport/mocks/mock_inhooks_config_service.go -package mocks
bin/mockgen -source pkg/services/message_builder.go -destination pkg/testsupport/mocks/mock_message_builder.go -package mocks
bin/mockgen -source pkg/services/message_enqueuer.go -destination pkg/testsupport/mocks/mock_message_enqueuer.go -package mocks