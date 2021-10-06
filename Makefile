# SPDX-License-Identifier: Unlicense

.PHONY: all clean lambda

all: clean lambda

clean:
	rm -f ov-lambda ov-lambda.zip

lambda:
	GOOS=linux go build -o ov-lambda
	zip ov-lambda.zip ov-lambda
	rm ov-lambda