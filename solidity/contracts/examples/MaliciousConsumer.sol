pragma solidity 0.4.24;


import "../Chainlinked.sol";


contract MaliciousConsumer is Chainlinked {
  uint256 constant private ORACLE_PAYMENT = 1 * LINK; // solium-disable-line zeppelin/no-arithmetic-operations

  constructor(address _link, address _oracle) public payable {
    setLinkToken(_link);
    setOracle(_oracle);
  }

  function () public payable {}

  function requestData(bytes32 _id, bytes _callbackFunc) public {
    ChainlinkLib.Run memory run = newRun(_id, this, bytes4(keccak256(_callbackFunc)));
    chainlinkRequest(run, ORACLE_PAYMENT);
  }

  function assertFail(bytes32, bytes32) public pure {
    assert(1 == 2);
  }

  function cancelRequestOnFulfill(bytes32 _requestId, bytes32) public {
    cancelChainlinkRequest(_requestId);
  }

  function remove() public {
    selfdestruct(address(0));
  }

  function stealEthCall(bytes32 _requestId, bytes32) public checkChainlinkFulfillment(_requestId) {
    // solium-disable-next-line security/no-call-value
    require(address(this).call.value(100)(), "Call failed");
  }

  function stealEthSend(bytes32 _requestId, bytes32) public checkChainlinkFulfillment(_requestId) {
    // solium-disable-next-line security/no-send
    require(address(this).send(100), "Send failed");
  }

  function stealEthTransfer(bytes32 _requestId, bytes32) public checkChainlinkFulfillment(_requestId) {
    address(this).transfer(100);
  }

  function doesNothing(bytes32, bytes32) public pure {}
}
