# Ellcrys PoW (Miner Package).

Ellcrys Proof-of-work is built on top of the Ethash formerly Dagger-Hashimoto algorithm.

Our proof of work is memory hard and Asic resistance which means miner can mine with their existince computer memory as against buying ASIC hardwares for that.

We made modifications to our implementation of ethash algorithm to prevent and also imporove the efficiency of our use case.


1. We opted for the use of blake2b hash function as against the original keccak and below are the reasons. 
			
			func hash256(dest []byte, data []byte) {
				hash := blake2b.Sum256(data)
				copy(dest, hash[:])
			}
			
			func hash512(dest []byte, data []byte) {
				hash := blake2b.Sum256(data)
				copy(dest, hash[:])
			}
		
	The two functions are available to hash data in the 256 algorithm i.e 32 bytes + 64 length size  and 512 algorithm i.e 64 bytes +  128 length size respectively.
	
		// Sequentially produce the initial dataset
		hash512(cache, seed)
		
	The above code uses the blake512 algorithm to generate the cache file which in turn is used to verify the nounce(submitted by the miner) by using the cache file to generate part of the DAG file needed for the operation.
	
	Reason for Changing the Hash function in Ethash
	
	* Changing the Ethash hash function from SHA-3(Keccak) to Blake2b to prevent the possibility of an existing mining  pool with significant hashrate on the Ethereum blockchain to easily launch a 51% attack on our blockchain(ellcrys).
	
	* Collision free/resistance
	
		Another reason we opted to use blake2b is the fact that it is collision free as at the time of writing this article and there are no known collision reported by any other company using this hash function
		
	* Secure
	
		Blake2b core compression function reuses the core function of ChaCha. and so far, this is one of the safest hashing fuction has rated on wikki and other cryptographic website

2. We generated our block hash which consist of the hash of (HashPrevBlock + HashMerkleRoot + Difficulty + Time) using the ASN.1 (Abstract Syntax Notation One) as against ethereum rlpHash.

	Asn1 implements parsing of DER-encoded ASN.1 data structures while RlpHash encodes specific data types (eg. strings, floats) according to higher-order protocols. 
	
		// HashNoNonce returns the hash of the block header which is used as input for the proof-of-work.
		func (h *Block) HashNoNonce() common.Hash {
			
			asn1B := asn1BlockNoNonce{
				HashPrevBlock:  h.HashPrevBlock,
				HashMerkleRoot: h.HashMerkleRoot,
				Difficulty:     h.Difficulty,
				Time: h.Time,
			}
			
			result, err := asn1.Marshal(asn1B)
			if err != nil {
				panic(err)
			}
			
			return sha256.Sum256(result)
		}
	
	as we can see in the code snippets, we use the asn1 to marshal the struct of thr block header before passing it to the hash function to get the digest.


3. Our difficulty calculation is based on the Homestead version of the ethereum leaving out Byzantuim and Frontier calculation.

4.  We updated our genesis block difficulty to 500000 so that it will be moderate for miners to work on.

		//GetGenesisDifficulty get the genesis block difficulty
		func (h *Block) GetGenesisDifficulty() *big.Int {
			return big.NewInt(500000)
		}

5. We created a structure for our block to using similar properties with all the blockchain blocks seen so far.

		type Block struct {
			Version        string   `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
			HashPrevBlock  string   `protobuf:"bytes,2,opt,name=hashPrevBlock,proto3" json:"hashPrevBlock,omitempty"`
			TX             []string `protobuf:"bytes,3,rep,name=tX" json:"tX,omitempty"`
			HashMerkleRoot string   `protobuf:"bytes,4,opt,name=hashMerkleRoot,proto3" json:"hashMerkleRoot,omitempty"`
			Time           string   `protobuf:"bytes,5,opt,name=time,proto3" json:"time,omitempty"`
			Nounce         uint64   `protobuf:"varint,6,opt,name=nounce,proto3" json:"nounce,omitempty"`
			Difficulty     string   `protobuf:"bytes,7,opt,name=difficulty,proto3" json:"difficulty,omitempty"`
			Number         uint64   `protobuf:"varint,8,opt,name=number,proto3" json:"number,omitempty"`
			PowHash        string   `protobuf:"bytes,9,opt,name=powHash,proto3" json:"powHash,omitempty"`
			PowResult      string   `protobuf:"bytes,10,opt,name=powResult,proto3" json:"powResult,omitempty"`
		}
