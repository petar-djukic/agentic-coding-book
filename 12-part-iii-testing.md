# Part III — Testing

*Outline stage.*

Testing code that no human wrote, and that no human will read line by
line. Differential testing against a reference implementation,
generated test suites and the circularity problem they create — tests
written by the same model, from the same assumptions, that produced the
code — and property-based testing as a way out of that circle.

The go-unix-utils work is the worked example: 107 utilities verified for
functional parity against the GNU reference binaries, which is a
verification method available whenever a reference implementation
exists, and a template for what to build when one does not.
