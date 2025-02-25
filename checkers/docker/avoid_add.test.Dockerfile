## These should be flagged

# <expect-error>
ADD ./source /destination

# <expect-error>
ADD https://example.com/file.tar.gz /destination/

# <expect-error>
ADD archive.tar.gz /extract-here/

# <expect-error>
ADD ["file1", "file2", "/dest/"]

# <expect-error>
ADD --chown=1000:1000 sourcefile /destination/

## These are safe and should not be flagged

# Using COPY instead of ADD
COPY ./source /destination

COPY ["file1", "file2", "/dest/"]

# Comments containing "ADD" should not trigger detection
# This is an example: ADD should not be used
