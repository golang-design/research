---
date: 2021-01-21T10:57:59+01:00
toc: true
slug: /zero-alloc-call-sched
tags:
  - Channel
  - EscapeAnalysis
  - GUI
  - MainThread
  - Thread
title: Scheduling Calls with Zero Allocation
draft: true
---

Author(s): [Changkun Ou](https://changkun.de)

GUI programming in Go is a little bit tricky. The infamous issue regarding interacting with the legacy GUI frameworks is that most of the graphics related APIs must be called from the main thread. This basically violates the concurrent nature of Go: A goroutine may be arbitrarily and randomly scheduled or rescheduled on different running threads, i.e., the same pice of code will be called from different threads over time, even without evolving the `go` keyword.

<!--more-->


## Background

TODO:

## The Main Thread

TODO:

## Cost Analysis and Optimization

TODO:

## Optimal Threading Control

TODO:

## Verification and Discussion

TODO:

## Conclusion

TODO:

## References

TODO: