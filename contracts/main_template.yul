object "Contract" {
    code {
        // TODO: Probably need to replace this
        // With wrapper made by hand just by adding few opcodes
        datacopy(0, dataoffset("runtime"), datasize("runtime"))
        return(0, datasize("runtime"))
    }
    object "runtime" {
        code {
            // Prerequisites: we always load 32 bytes from the stack
            // so we always need to clean them for our use case.

            // Check if called by owner, just replace it with your address
            if iszero(eq($owner, caller())) { revert(0, 0) }

            // Take first byte of the calldata by shifting it to the right with 0xf8 bits
            // calldata >> 248
            switch shr(0xf8, calldataload(0))
            // SwapIn
            case 0 {
                let pair := shr(0x60, calldataload(0x01))
                let amountIn := shr(sub(0x100, mul(8, sub(calldatasize(), 0x15))), calldataload(0x15))
                let reserveIn, reserveOut
                {
                    mstore(0, shl(0xe0, 0x0902f1ac))
                    if iszero(staticcall(gas(), pair, 0, 0x04, 0x04, 0x40)) { revert(0, 0) }
                    reserveIn:= mload(0x24)
                    reserveOut := mload(0x04)
                }
                let tokenIn
                {
                    mstore(0x44, shl(0xe0, 0xd21220a7))
                    if iszero(staticcall(gas(), pair, 0x44, 0x04, 0x48, 0x20)) { revert(0, 0) }
                    tokenIn := mload(0x48)
                }

                {
                    mstore(0x68, shl(0xe0, 0xa9059cbb)) mstore(add(0x68, 0x04), pair) mstore(add(0x68, 0x24), amountIn)
                    if iszero(call(gas(), tokenIn, 0, 0x68, 0x44, 0, 0))  { revert(0, 0) }
                }
                let aiwf := mul(amountIn, 9970)
                let v := div(mul(reserveOut, aiwf), add(mul(reserveIn,10000), aiwf))

                {
                    mstore(0x8c, shl(0xe0, 0x022c0d9f)) mstore(add(0x8c, 0x04), v)
                    mstore(add(0x8c, 0x44), address()) mstore(add(0x8c, 0x64), 0x80)
                    if iszero(call(gas(), pair, 0, 0x8c, 0xa4, 0, 0)) { revert(0, 0) }
                }
            }
            case 1 {
                let pair := shr(0x60, calldataload(0x01))
                let amountIn := shr(sub(0x100, mul(8, sub(calldatasize(), 0x15))), calldataload(0x15))
                let reserveIn, reserveOut
                {
                    mstore(0, shl(0xe0, 0x0902f1ac))
                    if iszero(staticcall(gas(), pair, 0, 0x04, 0x04, 0x40)) { revert(0, 0) }
                    reserveIn := mload(0x04)
                    reserveOut := mload(0x24)
                }
                let tokenIn
                {
                    mstore(0x44, shl(0xe0, 0x0dfe1681))
                    if iszero(staticcall(gas(), pair, 0x44, 0x04, 0x48, 0x20)) { revert(0, 0) }
                    tokenIn := mload(0x48)
                }

                {
                    mstore(0x68, shl(0xe0, 0xa9059cbb)) mstore(add(0x68, 0x04), pair) mstore(add(0x68, 0x24), amountIn)
                    if iszero(call(gas(), tokenIn, 0, 0x68, 0x44, 0, 0))  { revert(0, 0) }
                }
                let aiwf := mul(amountIn, 9970)
                let v := div(mul(reserveOut, aiwf), add(mul(reserveIn,10000), aiwf))

                {
                    mstore(0x8c, shl(0xe0, 0x022c0d9f)) mstore(add(0x8c, 0x24), v)
                    mstore(add(0x8c, 0x44), address()) mstore(add(0x8c, 0x64), 0x80)
                    if iszero(call(gas(), pair, 0, 0x8c, 0xa4, 0, 0)) { revert(0, 0) }
                }
            }
            // Withdraw ERC20
            case 2 {
            // 0x02
                let token := shr(0x60, calldataload(0x01))
                let contractBalance
                {
                    mstore(0, shl(0xe0, 0x70a08231)) mstore(add(0, 0x04), address())
                    if iszero(staticcall(gas(), token, 0, 0x24, 0x24, 0x20)) { revert(0, 0) }
                    contractBalance := mload(0x24)
                }

                if iszero(contractBalance) { revert(0, 0) }

                {
                    mstore(0x44, shl(0xe0, 0xa9059cbb))
                    mstore(add(0x44, 0x04), $owner) // TODO: place your address instead of 0xfefe
                    mstore(add(0x44, 0x24), sub(contractBalance, 0x01))
                    if iszero(call(gas(), token, 0, 0x44, 0x44, 0, 0)) { revert(0, 0) }
                }
            }
            case 3 {
                let to := shr(0x60, calldataload(0x01))
                mstore(0, to)
                return(0, 0x20)
            }
            default {
                revert(0, 0)
            }
        }
    }
}

