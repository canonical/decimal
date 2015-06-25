check: decimalTests
	go test ./...

decimalTests: add.decTest subtract.decTest multiply.decTest

%.decTest: dectest.zip
	unzip dectest.zip $@

dectest.zip:
	wget http://speleotrove.com/decimal/dectest.zip

clean:
	rm -f dectest.zip
	rm -f *decTest

