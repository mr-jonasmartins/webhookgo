build:
	GOOS=linux GOARCH=amd64 go build -o dist/webhookgo

send-exe-remote:
	rsync dist/webhookgo root@45.77.107.165:/root/linkdown/hook
	rsync index.html root@45.77.107.165:/root/linkdown/hook

send-service-remote:
	rsync webhookgo.service root@45.77.107.165:/root/linkdown/hook

deploy: send-exe-remote send-service-remote
	ssh -t root@45.77.107.165 '\
		mv /root/linkdown/hook/webhookgo.service /etc/systemd/system/ \
		&& systemctl enable webhookgo \
		&& systemctl restart webhookgo 