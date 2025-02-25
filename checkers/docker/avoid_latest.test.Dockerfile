## test_unsafe: These should be flagged

# <expect-error>
FROM ubuntu:latest

# <expect-error>
FROM node:latest

# <expect-error>
FROM python:latest

# <expect-error>
FROM alpine:latest AS builder


## test_safe: These are safe and should not be flagged

# Using a specific version
FROM ubuntu:20.04

FROM python:3.10

FROM alpine:3.17 AS builder

# Should not flag comments containing "latest"
# FROM image:latest
