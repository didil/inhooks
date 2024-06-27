rm -rf pkg/testsupport/mocks/mock_*.go

services=(
    "redis_store"
    "time_service"
    "inhooks_config_service"
    "message_builder"
    "message_enqueuer"
    "message_fetcher"
    "message_processor"
    "processing_results_service"
    "scheduler_service"
    "retry_calculator"
    "processing_recovery_service"
    "cleanup_service"
    "message_verifier"
    "payload_transformer"
)

for service in ${services[@]}
do
    bin/mockgen -source pkg/services/${service}.go -destination pkg/testsupport/mocks/mock_${service}.go -package mocks
done
