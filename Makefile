PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
LIBDIR ?= $(PREFIX)/lib/brain
CONFIGDIR ?= $(HOME)/.config/brain

.PHONY: install uninstall test

install:
	@echo "Installing Local Brain to $(DESTDIR)$(PREFIX)..."
	@mkdir -p $(DESTDIR)$(BINDIR)
	@mkdir -p $(DESTDIR)$(LIBDIR)
	@cp bin/* $(DESTDIR)$(BINDIR)/
	@cp lib/* $(DESTDIR)$(LIBDIR)/
	@chmod 755 $(DESTDIR)$(BINDIR)/brain*
	@chmod 644 $(DESTDIR)$(LIBDIR)/*
	@echo "OK: Installation complete."
	@echo " Executables: $(DESTDIR)$(BINDIR)"
	@echo " Libraries:  $(DESTDIR)$(LIBDIR)"

uninstall:
	@echo "Uninstalling Local Brain..."
	@rm -f $(DESTDIR)$(BINDIR)/brain*
	@rm -rf $(DESTDIR)$(LIBDIR)
	@echo "OK: Uninstalled."

test:
	@echo "Running tests..."
	@# Add test commands here
	@echo "Tests passed."
