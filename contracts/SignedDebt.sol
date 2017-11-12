pragma solidity ^0.4.0;

contract SignedDebt {
    address creator;
    address from;
    address to;
    string currency;
    uint value;
    uint fromSigned;
    uint toSigned;

    event Signed();

    function SignedDebt(string _currency, uint _value) public {
        require(_value > 0 && keccak256(_currency) != keccak256(""));
        creator = msg.sender;
        currency = _currency;
        value = _value;
    }

    function Currency() constant public returns(string) {
        return currency;
    }

    function Value() constant public returns(uint) {
        return value;
    }

    function Creator() constant public returns(address)  {
        return creator;
    }

    function From() constant public returns(address)  {
        return from;
    }

    function To() constant public returns(address) {
        return to;
    }

    function JoinFrom() public {
        require(from == address(0) && to != msg.sender);
        from = msg.sender;
    }

    function JoinTo() public {
        require(to == address(0) && from != msg.sender);
        to = msg.sender;
    }

    function Sign() public returns(bool) {
        require(from == msg.sender && fromSigned == 0 || to == msg.sender && toSigned == 0);
        if (msg.sender == from) {
            if (fromSigned == 0) {
                fromSigned = now;
                if (toSigned > 0) {
                    Signed();
                }
            }
            return toSigned > 0;
        }
        else if (msg.sender == to) {
            if (toSigned == 0) {
                toSigned = now;
                if (fromSigned > 0) {
                    Signed();
                }
            }
            return fromSigned > 0;
        }
    }

    function IsSigned() constant public returns(uint) {
        if (fromSigned == 0 || toSigned == 0) {
            return 0;
        } else if (fromSigned > toSigned) {
            return fromSigned;
        } else {
            return toSigned;
        }
    }
}