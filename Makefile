.PHONY: all

commands = migrate server queue

all: $(commands)

$(commands): %: cmd/%/main.go
	go build -o sn-$@ $<

clean:
	@rm -f deposit-*
