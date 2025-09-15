CC = gcc
CFLAGS = -Wall -Wextra -O2 -std=c99
LIBS = 

# Windows specific
ifeq ($(OS),Windows_NT)
    CFLAGS += -D_WIN32
    LIBS += -lkernel32
    TARGET = wipe-tool.exe
else
    TARGET = wipe-tool
endif

$(TARGET): wipe-tool.c
	$(CC) $(CFLAGS) -o $(TARGET) wipe-tool.c $(LIBS)

clean:
	rm -f $(TARGET)

install:
	cp $(TARGET) ../gui/

.PHONY: clean install