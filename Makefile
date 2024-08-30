build:
	GOOS=linux GOARCH=amd64 go build -o dist/webhookgo

send-exe-remote:
	rsync dist/webhookgo root@149.28.54.10:/var/www/html/webhookgo.finte.com.br
	rsync index.html root@149.28.54.10:/var/www/html/webhookgo.finte.com.br

send-service-remote:
	rsync webhookgo.service root@149.28.54.10:/var/www/html/webhookgo.finte.com.br

deploy: send-exe-remote send-service-remote
	ssh -t root@149.28.54.10 '\
		mv /var/www/html/webhookgo.com.br/webhookgo.service /etc/systemd/system/ \
		&& systemctl enable webhookgo \
		&& systemctl restart webhookgo \