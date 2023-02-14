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

The `varlog` service is a simple demonstration program.
See [Design Issues](#design-issues) below for a discussion of
how one might revise the program for production use.


# `/var/log` Service

# Logging

# Design Issues

## Resource Model

## Service API
