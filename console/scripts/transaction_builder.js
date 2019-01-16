/**
 * TXBuilder provides a class that wrappers the transaction
 * builder implemented in Go allows for regular lower-cased
 * method names inline with Javascript naming practice.
 * @param {Object} builder Native transaction builder implementation
 */
function TxBalanceBuilder(builder) {
	this.builder = builder;
}

TxBalanceBuilder.prototype.payload = function(finalize) {
	return this.builder.Payload(finalize);
};

TxBalanceBuilder.prototype.packed = function() {
	return this.builder.Packed();
};

TxBalanceBuilder.prototype.send = function() {
	return this.builder.Send();
};

TxBalanceBuilder.prototype.nonce = function(nonce) {
	this.builder.Nonce(nonce);
	return this;
};

TxBalanceBuilder.prototype.from = function(from) {
	this.builder.From(from);
	return this;
};

TxBalanceBuilder.prototype.senderPubKey = function(pk) {
	this.builder.SenderPubKey(pk);
	return this;
};

TxBalanceBuilder.prototype.to = function(addr) {
	this.builder.To(addr);
	return this;
};

TxBalanceBuilder.prototype.value = function(amount) {
	this.builder.Value(amount);
	return this;
};

TxBalanceBuilder.prototype.fee = function(amount) {
	this.builder.Fee(amount);
	return this;
};

TxBalanceBuilder.prototype.reset = function() {
	this.builder.Reset();
	return this;
};

// Add the builder class to the 'ell' namespace
ell["balance"] = function() {
	return new TxBalanceBuilder(_system.balance());
};
