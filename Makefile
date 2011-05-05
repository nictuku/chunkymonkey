SERVER_BINARY=chunkymonkey
EXTRA_BINARIES=intercept inspectlevel datatests whatmob
DIAGRAMS=diagrams/top-level-architecture.png

server: $(SERVER_BINARY)

all: server extra

extra: $(EXTRA_BINARIES)

clean:
	@-rm $(SERVER_BINARY) $(EXTRA_BINARIES)
	@gd -q -clean src

fmt:
	@gd -q -fmt --tab src

test: datatests
	@-rm -r .test_obj/tmp*
	@mkdir -p .test_obj
	@gd -q -L .test_obj -t src/lib
	@./datatests

libs:
	@gd -q src/lib

chunkymonkey: libs
	@gd -q -I src/lib -o $@ src/$@

intercept: libs
	@gd -q -I src/lib -o $@ src/$@

inspectlevel: libs
	@gd -q -I src/lib -o $@ src/$@

datatests: libs
	@gd -q -I src/lib -o $@ src/$@

whatmob: libs
	@gd -q -I src/lib -o $@ src/$@

docs: $(DIAGRAMS)

%.png: %.dot
	@dot -Tpng $< -o $@

.PHONY: all clean docs extra fmt libs server test
