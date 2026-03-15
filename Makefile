deploy-gui:
	rsync -avz --delete index.html manifest.json sw.js icon.svg icon-512.svg favicon.svg tom-rose.de:./httpdocs/herdbook

deploy-server:
	go build -o herdbook .
	rsync -avz --delete herdbook tom-rose.de:./herdbook/herdbook
