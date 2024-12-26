// SPDX-License-Identifier: MIT
pragma solidity ^0.8.23;

contract Inner {
    event TestEvent(uint256);

    function test() public returns (uint256) {
        emit TestEvent(42);
        return 42;
    }
}

contract BenchmarkInternalCall {
    Inner _inner;

    constructor() {
        _inner = new Inner();
    }

    function benchmarkInternalCall(uint256 iterations) public returns (uint256) {
        uint256 n = 0;

        for (uint256 i = 0; i < iterations; i++) {
            n += _inner.test();
        }

        return n;
    }
}
