#include <iostream>
#include <chrono>

void empty() {}

int main() {
    int n = 1000000;
    for (int j = 0; j < 10; j++) {
        std::chrono::nanoseconds since(0);
        for (int i = 0; i < n; i++) {
            auto start = std::chrono::steady_clock::now();
            empty();
            since += std::chrono::steady_clock::now() - start;
        }
        std::cout << "avg since: " << since.count() / n << "ns \n";
    }
}
