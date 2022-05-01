all:
	hugo
s:
	hugo server -D

pdf:
	go build pdfgen.go