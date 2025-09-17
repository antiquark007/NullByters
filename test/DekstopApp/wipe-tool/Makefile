CC = gcc
CFLAGS = -Wall -Wextra -O2
PYTHON = python3
PYTHON_SCRIPT = wipe.py

all: wipe-tool

wipe-tool: wipe-tool.c
	$(CC) $(CFLAGS) -o wipe-tool wipe-tool.c

run: wipe-tool
	./wipe-tool

clean:
	rm -f wipe-tool

install:
	$(PYTHON) -m pip install -r ../Cert_Tool/requirements.txt

.PHONY: all run clean install