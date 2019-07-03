/**
 * TXBuilder provides a class that wrappers the transaction
 * builder implemented in Go allows for regular lower-cased
 * method names inline with Javascript naming practice.
 * @param {Object} builder Native transaction builder implementation
 */
function TxTransferBuilder(builder) {
	this.builder = builder;
}

TxTransferBuilder.prototype.payload = function(finalize) {
	return this.builder.Payload(finalize);
};

TxTransferBuilder.prototype.serialize = function() {
	return this.builder.Serialize();
};

TxTransferBuilder.prototype.send = function() {
	return this.builder.Send();
};

TxTransferBuilder.prototype.nonce = function(nonce) {
	this.builder.Nonce(nonce);
	return this;
};

TxTransferBuilder.prototype.from = function(from) {
	this.builder.From(from);
	return this;
};

TxTransferBuilder.prototype.senderPubKey = function(pk) {
	this.builder.SenderPubKey(pk);
	return this;
};

TxTransferBuilder.prototype.to = function(addr) {
	this.builder.To(addr);
	return this;
};

TxTransferBuilder.prototype.value = function(amount) {
	this.builder.Value(amount);
	return this;
};

TxTransferBuilder.prototype.fee = function(amount) {
	this.builder.Fee(amount);
	return this;
};

TxTransferBuilder.prototype.reset = function() {
	this.builder.Reset();
	return this;
};

// Add the balance builder class to the 'ell' namespace
ell["balance"] = function() {
	return new TxTransferBuilder(_system.balance());
};

// Add ticket bid builder class to the 'ell' namespace
ell["ticketBid"] = function () {
	return new TxTransferBuilder(_system.ticketBid())
}