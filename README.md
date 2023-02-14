# varlog

# Table of Contents
* [Introduction](#introduction)
* [`/var/log` Service](#varlog-service)
* [Logging](#logging)
* [Design Issues](#design-issues)
  * [Resource Model](#resource-model)
  * [Service API](#service-api)

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
  Listing a directory gives entries under that directory.

The `varlog` service is a demonstration program.
See [Design Issues](#design-issues) below for a discussion of
how one might revise the program for production use.

# `/var/log` Service

The service of this demonstration package provides two HTTP request endpoints,
as summarized above.
This section provides the details of each endpoint.

* `read`
  * Operation.  This endpoint opens a given file within `/var/log`,
    applies an optional text filter to match (or drop) lines,
    and presents the most recent entries first, up to a given count.
  * HTTP Method: `GET`
  * Path: `/read`
  * Query Parameters
    * `name=`_filepath_ \
      Required.
      Specifies the file to be read.  The _filepath_ value is used to construct
      the full path name as `/var/log/`_filepath_.
      This "file" must be a regular file---not a directory, a symbolic link,
      nor a special file of any kind.
      Note that _filepath_ can contain multiple levels, giving full access to the
      `/var/log` directory tree.  For example, if _filepath_ has the value
      `dir1/dir2/file-abc`, the full path to be read is `/var/log/dir1/dir2/file-abc`.
    * `filter=`_text_ \
      `filter=-`_text_ \
      Optional.
      If present, specifies an exact text pattern that to apply to lines in the file.
      The positive form, `filter=`_text_, requires _text_ to be present;
      lines without the pattern are omitted from the response.
      The negative form, `filter=-`_text_, requires _text_ NOT to be present;
      lines with the pattern are omitted from the response.
      If this parameter is not present, the filter allows all lines in the file
      to be part of the response.
    * `count=`_number_ \
      Optional.
      If present, specifies the maximum line count for the response body.
      The _count_ most recent, filtered lines are selected from the file.
      If the `filter` parameter disqualifies a line, it does _not_ count
      against this `count` limit.
      If this parameter is not present, all qualifying lines appear
      in the response body.
  * Error conditions.
    HTTP status codes in the 400 and 500 range indicate error conditions.
    Consult [List of HTTP status codes](
	    https://en.wikipedia.org/wiki/List_of_HTTP_status_codes
    ) or similar references for details.


# Logging

# Design Issues

## Resource Model

## Service API

## Authentication
