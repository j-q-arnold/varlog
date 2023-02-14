# varlog

# Table of Contents
* [Introduction](#introduction)
* [`/var/log` Service](#varlog-service)
  * [Running the Service](#running-the-service)
* [`/var/log` Client](#varlog-client)
* [Logging](#logging)
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
  Listing a directory gives entries under that directory.

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
      The _path_ value may not be empty.
      `/var/log` directory tree.  For example, if _path_ has the value
      `dir1/dir2/file-abc`, the full path to be read is `/var/log/dir1/dir2/file-abc`.
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
      Specifies the entry to be read.  The _path_ value is used to construct
      the full path name as `/var/log/`_path_.
      If this `name` parameter is empty or not present, the base directory
      `/var/log` is used as the full path name.
      If the resulting entry is a directory, that directory is read
      and all qualifying children are added to the response.
      If the resulting entry is a regular file, that regular file itself
      appears as the single entry in the response.
      Note that _path_ can contain multiple levels, giving full access to the
      `/var/log` directory tree.  For example, if _path_ has the value
      `dir1/dir2/file-abc`, the full path to be read is `/var/log/dir1/dir2/file-abc`.
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
      `/var/log/dir/file`, the key's value would be `dir/file`.
    * `"type"`.  This key's value indicates the entry type: `"file"` for a
      regular file and `"dir"` for a directory.
      Other types of entries are omitted from the response.
  * Error conditions.
    HTTP status codes in the 400 and 500 range indicate error conditions.
    Consult [List of HTTP status codes](
	    https://en.wikipedia.org/wiki/List_of_HTTP_status_codes
    ) or similar references for details.

## Running the Service

# `/var/log` Client

A web browser can be used to exercise the service.
Some example addresses follow (running on the same
machine as the service):

* `http://localhost:8000/list?filter=syslog` \
  List all the files and directories directly under `/var/log`
  that have `syslog` in the name.
* `http://localhost:8000/read?name=syslog-saturn-2023-01-31&filter=ERROR&count=100` \
  Read the 100 latest `ERROR` messages from `syslog-saturn-2023-01-31`.
* `http://localhost:8000/read?name=syslog-saturn-2023-01-31&filter=-INFO&count=100` \
  Similar to the previous example, except this allows lines _except_
  the `INFO` entries.


# Logging

# Design Issues

## Resource Model

## Service API

## Authentication

## Observability

## Build & Deployment

## Performance Enhancements
* Index file, record lines, update cache
* compress response body for "large" responses
* chunking
