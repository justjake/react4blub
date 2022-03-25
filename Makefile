BUILD=build
DEBUG=$(BUILD)/debug
RELEASE=$(BUILD)/release
CFLAGS=-Wall
CFLAGS_DEBUG=$(CFLAGS)
CFLAGS_RELEASE=$(CFLAGS) -ftlo -O4

$(RELEASE)/%.o: %.c
	$(CC) $(CFLAGS_RELEASE) -o $@ $< 

$(DEBUG)/%.o: %.c
	$(CC) $(CFLAGS_DEBUG) -o $@ $< 

$(DEBUG)/test.exe: $(DEBUG)/sqlite.o $(DEBUG)/test.o
	$(CC) $* -o $@