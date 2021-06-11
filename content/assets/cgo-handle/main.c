// Copyright 2021 The golang.design Initiative Authors.
// All rights reserved. Use of this source code is governed
// by a MIT license that can be found in the LICENSE file.
//
// Written by Changkun Ou <changkun.de>

#include <stdint.h>

extern void GoFunc(uintptr_t handle);

void cFunc(uintptr_t handle) {
    GoFunc(handle);
}