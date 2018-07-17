package testdata

var ProcessDotGoJSON = []string{
	`{
        "header": {
            "parentHash": null,
            "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
            "number": 1,
            "stateRoot": "abc",
            "transactionsRoot": "jsjhf9e3i9nfi",
            "nonce": 0,
            "mixHash": "",
            "difficulty": "500000",
            "timestamp": 0
        },
        "transactions": [
        ],
        "hash": "hash_1",
        "sig": "abc"
    }`,
	`{
        "header": {
            "parentHash": "hash_1",
            "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
            "number": 1,
            "stateRoot": "abc",
            "transactionsRoot": "jsjhf9e3i9nfi",
            "nonce": 0,
            "mixHash": "",
            "difficulty": "500000",
            "timestamp": 0
        },
        "transactions": [
            {
                "type": 1,
                "nonce": 2,
                "to": "efjshfhh389djn29snmnvuis",
                "senderPubKey": "xsj2909jfhhjskmj99k",
                "from": "jxzzf1HtQL7M9Eh792iGHTvb6fsnnPipas",
                "value": "100.333",
                "timestamp": 1508673895,
                "fee": "0.00003",
                "invokeArgs": null,
                "sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
                "hash": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs"
            }
        ],
        "hash": "hash_2",
        "sig": "abc"
    }`,
	`{
        "header": {
            "parentHash": "hash_2",
            "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
            "number": 1,
            "stateRoot": "abc",
            "transactionsRoot": "jsjhf9e3i9nfi",
            "nonce": 0,
            "mixHash": "",
            "difficulty": "500000",
            "timestamp": 0
        },
        "transactions": [
            {
                "type": 1,
                "nonce": 2,
                "to": "efjshfhh389djn29snmnvuis",
                "senderPubKey": "xsj2909jfhhjskmj99k",
                "from": "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
                "value": "100.333",
                "timestamp": 1508673895,
                "fee": "0.00003",
                "invokeArgs": null,
                "sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
                "hash": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs"
            }
        ],
        "hash": "hash_3",
        "sig": "abc"
    }`,
	`{
        "header": {
            "parentHash": "hash_3",
            "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
            "number": 1,
            "stateRoot": "abc",
            "transactionsRoot": "jsjhf9e3i9nfi",
            "nonce": 0,
            "mixHash": "",
            "difficulty": "500000",
            "timestamp": 0
        },
        "transactions": [
            {
                "type": 1,
                "nonce": 2,
                "to": "efjshfhh389djn29snmnvuis",
                "senderPubKey": "xsj2909jfhhjskmj99k",
                "from": "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
                "value": "100_333",
                "timestamp": 1508673895,
                "fee": "0.00003",
                "invokeArgs": null,
                "sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
                "hash": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs"
            }
        ],
        "hash": "hash_4",
        "sig": "abc"
    }`,
	`{
        "header": {
            "parentHash": "hash_4",
            "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
            "number": 1,
            "stateRoot": "abc",
            "transactionsRoot": "jsjhf9e3i9nfi",
            "nonce": 0,
            "mixHash": "",
            "difficulty": "500000",
            "timestamp": 0
        },
        "transactions": [
            {
                "type": 1,
                "nonce": 2,
                "to": "efjshfhh389djn29snmnvuis",
                "senderPubKey": "xsj2909jfhhjskmj99k",
                "from": "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
                "value": "1",
                "timestamp": 1508673895,
                "fee": "0.00003",
                "invokeArgs": null,
                "sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
                "hash": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs"
            }
        ],
        "hash": "hash_5",
        "sig": "abc"
    }`,
	`{
        "header": {
            "parentHash": "hash_5",
            "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
            "number": 1,
            "stateRoot": "abc",
            "transactionsRoot": "jsjhf9e3i9nfi",
            "nonce": 0,
            "mixHash": "",
            "difficulty": "500000",
            "timestamp": 0
        },
        "transactions": [
            {
                "type": 1,
                "nonce": 2,
                "to": "eQ9TnvMUUsB8ztZchSe3o7f5XfifEmZvJR",
                "senderPubKey": "xsj2909jfhhjskmj99k",
                "from": "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
                "value": "1",
                "timestamp": 1508673895,
                "fee": "0.00003",
                "invokeArgs": null,
                "sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
                "hash": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs"
            }
        ],
        "hash": "hash_6",
        "sig": "abc"
    }`,
	`{
        "header": {
            "parentHash": "hash_6",
            "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
            "number": 1,
            "stateRoot": "abc",
            "transactionsRoot": "jsjhf9e3i9nfi",
            "nonce": 0,
            "mixHash": "",
            "difficulty": "500000",
            "timestamp": 0
        },
        "transactions": [
            {
                "type": 1,
                "nonce": 2,
                "to": "eQ9TnvMUUsB8ztZchSe3o7f5XfifEmZvJR",
                "senderPubKey": "xsj2909jfhhjskmj99k",
                "from": "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
                "value": "1",
                "timestamp": 1508673895,
                "fee": "0.00003",
                "invokeArgs": null,
                "sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
                "hash": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs"
            },
            {
                "type": 1,
                "nonce": 3,
                "to": "eQ9TnvMUUsB8ztZchSe3o7f5XfifEmZvJR",
                "senderPubKey": "xsj2909jfhhjskmj99k",
                "from": "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
                "value": "1",
                "timestamp": 1508673896,
                "fee": "0.00003",
                "invokeArgs": null,
                "sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
                "hash": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs"
            }
        ],
        "hash": "hash_7",
        "sig": "abc"
    }`,
	`{
        "header": {
            "parentHash": "hash_7",
            "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
            "number": 1,
            "stateRoot": "abc",
            "transactionsRoot": "jsjhf9e3i9nfi",
            "nonce": 0,
            "mixHash": "",
            "difficulty": "500000",
            "timestamp": 0
        },
        "transactions": [
            {
                "type": 1,
                "nonce": 2,
                "to": "efjshfhh389djn29snmnvuis",
                "senderPubKey": "xsj2909jfhhjskmj99k",
                "from": "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
                "value": "1",
                "timestamp": 1508673895,
                "fee": "0.00003",
                "invokeArgs": null,
                "sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
                "hash": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs"
            },
            {
                "type": 1,
                "nonce": 3,
                "to": "efjshfhh389djn29snmnvuis",
                "senderPubKey": "xsj2909jfhhjskmj99k",
                "from": "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
                "value": "1",
                "timestamp": 1508673896,
                "fee": "0.00003",
                "invokeArgs": null,
                "sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
                "hash": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs"
            }
        ],
        "hash": "hash_8",
        "sig": "abc"
    }`,
}

var ProcessStaleOrInvalidBlockData = []string{
	`{
        "header": {
            "parentHash": null,
            "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
            "number": 1,
            "stateRoot": "abc",
            "transactionsRoot": "jsjhf9e3i9nfi",
            "nonce": 0,
            "mixHash": "",
            "difficulty": "500000",
            "timestamp": 0
        },
        "transactions": [
        ],
        "hash": "hash_1",
        "sig": "abc"
    }`,
	`{
        "header": {
            "parentHash": "hash_1",
            "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
            "number": 2,
            "stateRoot": "abc",
            "transactionsRoot": "jsjhf9e3i9nfi",
            "nonce": 0,
            "mixHash": "",
            "difficulty": "500000",
            "timestamp": 0
        },
        "transactions": [
        ],
        "hash": "hash_2",
        "sig": "abc"
    }`,
	`{
        "header": {
            "parentHash": "hash_2",
            "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
            "number": 3,
            "stateRoot": "abc",
            "transactionsRoot": "jsjhf9e3i9nfi",
            "nonce": 0,
            "mixHash": "",
            "difficulty": "500000",
            "timestamp": 0
        },
        "transactions": [
        ],
        "hash": "hash_3",
        "sig": "abc"
    }`,
	`{
        "header": {
            "parentHash": "hash_1",
            "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
            "number": 2,
            "stateRoot": "abc",
            "transactionsRoot": "jsjhf9e3i9nfi",
            "nonce": 0,
            "mixHash": "",
            "difficulty": "500000",
            "timestamp": 0
        },
        "transactions": [
        ],
        "hash": "hash_4",
        "sig": "abc"
    }`,
	`{
        "header": {
            "parentHash": "hash_2",
            "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
            "number": 3,
            "stateRoot": "abc",
            "transactionsRoot": "ahshgde3i9nfi",
            "nonce": 0,
            "mixHash": "",
            "difficulty": "500000",
            "timestamp": 0
        },
        "transactions": [
        ],
        "hash": "hash_4",
        "sig": "abc"
    }`,
	`{
        "header": {
            "parentHash": "hash_2",
            "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
            "number": 5,
            "stateRoot": "abc",
            "transactionsRoot": "ahshgde3i9nfi",
            "nonce": 0,
            "mixHash": "",
            "difficulty": "500000",
            "timestamp": 0
        },
        "transactions": [
        ],
        "hash": "hash_4",
        "sig": "abc"
    }`,
}

var ProcessStateRootData = []string{
	`{
        "header": {
            "parentHash": "hash_2",
            "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
            "number": 5,
            "stateRoot": "abc",
            "transactionsRoot": "ahshgde3i9nfi",
            "nonce": 0,
            "mixHash": "",
            "difficulty": "500000",
            "timestamp": 0
        },
        "transactions": [
        ],
        "hash": "hash_4",
        "sig": "abc"
    }`,
	`{
        "header": {
            "parentHash": null,
            "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
            "number": 1,
            "stateRoot": "abc",
            "transactionsRoot": "ahshgde3i9nfi",
            "nonce": 0,
            "mixHash": "",
            "difficulty": "500000",
            "timestamp": 0
        },
        "transactions": [
        ],
        "hash": "hash_1",
        "sig": "abc"
    }`,
	`{
        "header": {
            "parentHash": "hash_1",
            "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
            "number": 2,
            "stateRoot": "abc",
            "transactionsRoot": "ahshgde3i9nfi",
            "nonce": 0,
            "mixHash": "",
            "difficulty": "500000",
            "timestamp": 0
        },
        "transactions": [
            {
                "type": 1,
                "nonce": 3,
                "to": "",
                "senderPubKey": "",
                "from": "",
                "value": "10",
                "timestamp": 1508673896,
                "fee": "0.00003",
                "invokeArgs": null,
                "sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
                "hash": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs"
            }
        ],
        "hash": "hash_2",
        "sig": "abc"
    }`,
}

var ProcessMockBlockData = []string{
	`{
        "header": {
            "parentHash": null,
            "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
            "number": 1,
            "stateRoot": "abc",
            "transactionsRoot": "ahshgde3i9nfi",
            "nonce": 0,
            "mixHash": "",
            "difficulty": "500000",
            "timestamp": 0
        },
        "transactions": [
            {
                "type": 1,
                "nonce": 1,
                "to": "",
                "senderPubKey": "",
                "from": "",
                "value": "1",
                "timestamp": 1508673896,
                "fee": "0.00003",
                "invokeArgs": null,
                "sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
                "hash": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs"
            }
        ],
        "hash": "hash_1",
        "sig": "abc"
    }`,
}
