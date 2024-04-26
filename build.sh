pnpm build && atlas schema apply --auto-approve --to "file://schema.sql" --url "sqlite://data.db" --dev-url "sqlite://dev?mode=memory" && sqlc generate && templ generate && go build -o ./bin/gssc-go-$(date "+%d-%m-%y") .
rsync -avh /home/viktor/dev/go/gssc-go/bin/ /mnt/nvme/file_browser/srv/gssc-go/
