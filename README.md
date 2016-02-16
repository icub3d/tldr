# tldr

Program tldr is a program that prints out simplified man pages loosely based on cheat.

# Installation

If you have Go installed and your _$PATH_ include _$GOBIN_, simply:

	go get -u github.com/icub3d

Alternatively, you can download a linux x86_64 version from the release page.

# Usage

Try:

	tldr -h

for details, but essentially, you need to pull the tldrs by:

	tldr -pull

and then you can ask for them with:

	tldr [name]

For example:

	tldr tar


