# varlog

# Table of Contents
* [Introduction](#introduction)
* [`/var/log` Service](#varlog-service)
  * [Building and Running the Service](#building-and-running-the-service)
  * [Command Line Options](#command-line-options)
* [`/var/log` Client](#varlog-client)
* [Logging](#logging)
* [Test Data](#test-data)
* [Design Issues](#design-issues)
  * [Resource Model](#resource-model)
  * [Service API](#service-api)
  * [Authentication](#authentication)
  * [Observability](#observability)
  * [Build & Deployment](#build-deployment)
  * [Performance Enhancements](#performanc-enhancements)

# Introduction
This project provides a simple service to retrieve `/var/log` entries.
Given a file or directory under `/var/log`, two service endpoints
act on the given name:
* `read`: Open the file, apply an optional filter and line count,
  and return the specified number of lines, most recent first.
  Only files may be read; directories and special files are not allowed.
* `list`: Given a file or directory, list the children in the manner
  of the Unix `ls` command (sorted with file metadata).
  Listing a file gives the file itself.
  Listing a directory gives entries directly under that directory.

The `varlog` service is a demonstration program.
See [Design Issues](#design-issues) below for a discussion of
how one might revise the program for production use.

# `/var/log` Service

The service of this demonstration package provides two HTTP request endpoints,
as summarized above.
This section provides the details of each endpoint.

* `read`
  * Operation.  This endpoint opens a given file within `/var/log` for reading,
    applies an optional text filter to match (or drop) lines,
    and presents the most recent entries first, up to a given count.
  * HTTP Method: `GET`
  * URL Path: `/read`
  * Query Parameters
    * `name=`_path_ \
      Required.
      Specifies the file to be read.  The _path_ value is used to construct
      the full path name as `/var/log/`_path_.
      This "file" must be a regular file---not a directory, a symbolic link,
      nor a special file of any kind.
      Note that _path_ can contain multiple levels, giving full access to the
      `/var/log` directory tree.  For example, if _path_ has the value
      `dir1/dir2/file-abc`, the full path to be read is `/var/log/dir1/dir2/file-abc`.
      The _path_ value may not be empty, and it may not use `..`
      to escape the `/var/log` tree.
    * `filter=`_text_ \
      `filter=`_-text_ \
      Optional.
      If present and non-empty, specifies an exact text pattern to apply
      to lines in the file.
      The positive form, `filter=`_text_, requires _text_ to be present somewhere
      in the line;
      lines without the pattern are omitted from the response.
      The negative form, `filter=`_-text_, requires _text_ NOT to be present;
      lines with the pattern are omitted from the response.
      If this parameter is empty or not present, the filter allows all lines
      in the file to be part of the response.
      Note that filtering requires an exact match on _text_: no regular
      expression matching is applied.
    * `count=`_number_ \
      Optional.
      If present and positive, specifies the maximum line count for the response body.
      The _count_ most recent, filtered lines are selected from the file.
      If the `filter` parameter disqualifies a line, it does _not_ count
      against this limit.
      If this parameter is non-positive or not present, all qualifying
      lines appear in the response body.
    * `content-disposition=`_value_ \
      Optional.
      This specifies how to prepare the output:
      to show `inline` or to save as a download `attachment`.
      If omitted or empty, the server decides, based on the expected
      size of the results.  Small results are shown inline; large results
      are downloaded.
  * Response.
    The body of the response contains the selected lines, one line from
    the file per line in the response.
    As mentioned, the response lines appear most recent first.
  * Error conditions.
    HTTP status codes in the 400 and 500 range indicate error conditions.
    Consult [List of HTTP status codes](
	    https://en.wikipedia.org/wiki/List_of_HTTP_status_codes
    ) or similar references for details.

* `list`
  * Operation.  This endpoint examines a given directory
    (or file) within `/var/log`,
    finds the file and directory children, gathers certain metadata
    about the entries, and returns a JSON response to the client.
  * HTTP Method: `GET`
  * URL Path: `/list`
  * Query Parameters
    * `name=`_path_ \
      Optional.
      Specifies the entry to be listed.  The _path_ value is used to construct
      the full path name as `/var/log/`_path_.
      If this `name` parameter is empty or not present, the base directory
      `/var/log` is used as the full path name.
      If the resulting entry is a directory, that directory is read
      and all qualifying children are added to the response.
      If the resulting entry is a regular file, that regular file itself
      appears as the single entry in the response.
      Note that _path_ can contain multiple levels, giving full access to the
      `/var/log` directory tree.  For example, if _path_ has the value
      `dir1/dir2/file-abc`, the full path to be listed is `/var/log/dir1/dir2/file-abc`.
      The _path_ value may not use `..` to escape the `/var/log` tree.
    * `filter=`_text_ \
      `filter=`_-text_ \
      Optional.
      If present and non-empty, specifies an exact text pattern that to apply
      to response items.
      The positive form, `filter=`_text_, requires _text_ to be present
      in the name of an item;
      lines without the pattern are omitted from the response.
      The negative form, `filter=`_-text_, requires _text_ NOT to be present;
      entries with the pattern are omitted from the response.
      If this parameter is empty or not present, the filter allows all entries
      in the directory (file) to be part of the response.
  * Response.
    The response is a JSON array of objects.
    The response array can be empty, such as when a directory has no children.
    Response objects have the following name/value pairs.
    * `"name"`.  This key's value gives the name of the entry, relative to
      `/var/log`.  For example, if the full path of an entry is
      `/var/log/dir/file`, the key's value would be `"dir/file"`.
    * `"type"`.  This key's value indicates the entry type: `"file"` for a
      regular file and `"dir"` for a directory.
      Other types of entries are omitted from the response.
  * Error conditions.
    HTTP status codes in the 400 and 500 range indicate error conditions.
    Consult [List of HTTP status codes](
	    https://en.wikipedia.org/wiki/List_of_HTTP_status_codes
    ) or similar references for details.

## Building and Running the Service
This does not have a fully developed project.
These instructions assume Go is installed, and you
have initialized the Go environment variables.
Here are the minimal steps.
* The following commands use `$REPO` as an environment
  variable holding the path to the `varlog` repository.
  You'll need to adjust the following for your machine.
  ```
  $ export REPO=$HOME/varlog
  ```

* Change directory into the git repository.
  ```
  $ cd $REPO/service/varlog-srv
  $ go build .
  ```
  This builds the executable: `varlog-srv` (or `varlog-srv.exe` for Windows).

* Run the program.  This defaults to listening on port 8000, but you
  can change that if another server is listening there.
  ```
  $ ./varlog-srv -help	# to see a usage message
  $ ./varlog-srv	# to start the service
  ```

* If you are running on Windows or want to use local log files,
  redirect the root to local test data:
  ```
  $ ./varlog-srv -root $REPO/testdata/var/log
  ```

## Command Line Options
The server has a few command line options that control its behavior.
The default configuration would work on a typical linux machine,
but the following change behavior in useful ways for testing and development.

* `-port NUMBER` \
  Sets the port on which the server listens.
  Default is 8000, but this might be busy on some machines.
* `-root PATH` \
  Sets the root for the log file directory.
  This was shown above to use test data in the repository.
  Having only the real `/var/log` for test input is not satisfactory.
* `-chunk SIZE` \
  The service assumes some log files might be too big to read into memory.
  It thus reads log files in chunks, starting at the end of the file.
  Handling lines that cross chunk boundaries is the primary complication
  of the code.  This parameter controls that internal chunk size to test
  boundary conditions.  It is possible to run the service with `-chunk=1`,
  though that would be inadvisable for production.
  In this edge case condition, every line spans a chunk boundary,
  and every edge condition is forced to happen.

  Try running the service with various chunk sizes.  The behavior
  should be identical, regardless of the current size.

# `/var/log` Client

A web browser can be used to exercise the service.
Some example addresses follow, assuming the browser runs
on the same machine as the service.
This also assume you have started the service as above,
using the repository's test data.
See [Test Data](#test-data) below for details.

* `http://localhost:8000/list` \
  List all the files and directories directly under `/var/log`.
* `http://localhost:8000/list?filter=log` \
  List all the files and directories directly under `/var/log`
  that have `log` in the name.
* `http://localhost:8000/list?filter=-log` \
  List all the files and directories directly under `/var/log`
  that do not have `log` in the name.
* `http://localhost:8000/read?name=log-100&filter=ERROR&count=10` \
  Read the 10 latest `ERROR` messages from `log-100`.
* `http://localhost:8000/read?name=log-100&filter=-INFO&count=10` \
  Similar to the previous example, except this allows lines _except_
  the `INFO` entries.
* `http://localhost:8000/read?name=log-1M&count=10` \
  In this example, `log-1M` has 1 million lines (75MB file size).
  This caps the line count at 10 and results stream to the browser.
* `http://localhost:8000/read?name=log-1M` \
  This example also uses `log-1M` but requests the entire file.
  The server automatically applies a `Content-Disposition` header to
  download a file instead of displaying inline.


# Logging
Log messages are written to standard error for this program.
* `ERROR`: used for internal errors, things that should not happen.
* `WARNING`: Anything that could be caused by the client request:
  invalid name, bad parameter value, etc.
* `INFO`: Normal activity logging by the application.
* `DEBUG`: Temporary notes or other messages.


# Test Data
The repository has some test files that can be used.
Typical lines look like the following:
```
2023/02/16 07:40:46 aaaaa          0   DEBUG abcde fghij klmno pqrst uvwxy
2023/02/16 07:40:46 bbbbb          1    INFO abcde fghij klmno pqrst uvwxy
2023/02/16 07:40:46 ccccc          2 WARNING abcde fghij klmno pqrst uvwxy
2023/02/16 07:40:46 ddddd          3   ERROR abcde fghij klmno pqrst uvwxy
2023/02/16 07:40:46 eeeee          4   DEBUG abcde fghij klmno pqrst uvwxy
```

These are made to resemble actual log files, with some tweaks:
* All lines are the same length.  This is not realistic, but any
  problems with the output are easy to spot.
* Application tags, `aaaaa` and such, can be filtered.  One might
  Want to retrieve only lines related to a specific application.
* Every line in the file has a unique sequence number, starting at zero.
  When filtering lines or counting, one can use the sequence number to
  confirm data are as expected.
* Message levels, `ERROR` and such, also are useful filters.

A list of the useful files and brief descriptions.

* `log-0`: An empty file.
* `log-10`: A file with 10 lines
* `log-100`: A file with 100 lines
* `log-1K`: A file with 1000 lines
* `log-10K`: A file with 10,000 lines
* `log-1M`: A file with 1,000,000 lines.
* `log-nl`: A file with 14 lines, some empty, to use for
  "chunking" exercises.  More on this below.

A few interesting things one can try with these files.
* Filter on the application: `aaaaa`, `bbbbb`, etc.
* Filter on the message level: `INFO`, `ERROR`, etc.
* Try positive and negative filters.
* Use `count=n` to limit the lines returned.
* Filter on the sequence number, such as `filter=0000`,
  to see one of every 10,000 lines.

A final note about file "chunking".
Reading a file backwards presents a challenge when a line
spans two chunks.
The line suffix is read first (chunk N), the line prefix
is read next (chunk N-1), and the pieces must be spliced together
to form a complete line.
There are further complications for empty lines,
lines that span multiple chunks,
or when a newline is exactly at the end or beginning of a chunk
(or both).
Depending on the line length and the chunk size, these conditions
can be hard to enumerate and to test.

Thus the `-chunk` parameter lets one adjust the internal behavior
to suit the test data in hand.
A useful trick is to set `-chunk=1`, ensuring even "simple" input
files will exercise interesting boundary conditions.
The `log-nl` file has only a few lines, but they include empty
lines, consecutive empty lines, and empty lines at the start
and end of the file.
If the line reversal code had defects, `-chunk=1` would help expose them.

Streaming this `log-nl` to the browser makes it difficult to
see the empty lines.
One can use the URL query parameter, `content-disposition=attachment`,
to force a download.
Then one can open the resulting file directly.

# Design Issues

One could design a resource model to mirror `/var/log` (or a
file system in general).
That was not the problem here, but it could provide a more
general API for program-to-program communication.

An extended service model also would support other HTTP methods.
Services often use messages such as PUT/POST with request bodies,
allowing JSON flexibility in addition to query parameters.

An actual service also might need authentication, depending
on the network routing and service visibility.

## Observability
A production system should provide monitoring metrics.
Some of this could be standard kubernetes health check probes.
Individual requests should provide a context identifier,
tagging log entries to enable start-to-finish tracing.

## Build & Deployment
The build here is rudimentary.
Integrated into a team's build structure, github activity
would trigger automatic builds, run tests, push images to
container registries, etc.

## Performance Enhancements
First, any performance work should measure the service
and find any bottlenecks.
Here are a few ideas of what might happen and how one
might address those concerns.

* Compress the response bodies.
  Depending on the frequency of large responses, one might
  compress the response bodies.  This is not likely to help
  if 99% of all requests use small line counts, but it might
  work for large responses.
* Too many accesses to the same data.
  The system has some overhead to read and reverse lines.
  If "typical" files changed infrequently, one might consider
  indexing or caching the reversed data, reusing reversed
  files multiple times.  (Typical log files would change
  frequently, so this is not likely to be a real possibility.)
* File system issues.  One could increase (or decrease) the internal
  "chunk" size to reduce file system overhead.
