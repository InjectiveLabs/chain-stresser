
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.23;

contract Counter {
    uint256 private _count;

    event Increased(address indexed sender, uint256 newValue);
    event Decreased(address indexed sender, uint256 newValue);

    constructor() {
        _count = 0;
    }

    function increase() public {
        _count += 1;
        emit Increased(msg.sender, _count);
    }

    function decrease() public {
        require(_count > 0, "Count cannot be negative");
        _count -= 1;
        emit Decreased(msg.sender, _count);
    }

    function getCount() public view returns (uint256) {
        return _count;
    }
}
