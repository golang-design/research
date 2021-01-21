// Copyright (c) 2020 The golang.design Initiative Authors.
// All rights reserved.
//
// The code below is produced by Changkun Ou <hi@changkun.de>.

#include <iostream>
#include <chrono>

void target() {}

int main() {
    int n = 1000000;
    for (int j = 0; j < 10; j++) {
        std::chrono::nanoseconds since(0);
        for (int i = 0; i < n; i++) {
            auto start = std::chrono::steady_clock::now();
            target();
            since += std::chrono::steady_clock::now() - start;
        }

        auto overhead = -(std::chrono::steady_clock::now() -
                          std::chrono::steady_clock::now());
        since -= overhead * n;

        std::cout << "avg since: " << since.count() / n << "ns \n";
    }
}