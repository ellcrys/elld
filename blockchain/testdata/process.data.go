package testdata

var ProcessDotGoJSON = []string{
	`{
        "header": {
            "creatorPubKey": "48s9G48LD5eo5YMjJWmRjPaoDZJRNTuiscHMov6zDGMEqUg4vbG",
            "number": 1,
            "transactionsRoot": "0x8f845ec4453d2446695d5908cba62574850f97450fd7436421047101d88a58bb",
            "stateRoot": "0x8f845ec4453d2446695d5908cba62574850f97450fd7436421047101d88a58bb",
            "nonce": 1030,
            "difficulty": "2747646837",
            "timestamp": 1532623286,
            "mixHash": "0x8f845ec4453d2446695d5908cba62574850f97450fd7436421047101d88a58bb"
        },
        "transactions": [
            {
                "type": 1,
                "to": "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
                "from": "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
                "senderPubKey": "48s9G48LD5eo5YMjJWmRjPaoDZJRNTuiscHMov6zDGMEqUg4vbG",
                "value": "1",
                "Timestamp": 1532623441,
                "Fee": "0.1",
                "sig": "0x849bbe5877507cce082ca536445297f28679994d412973f32f8aa8fff68691e29b9189ff7b4999511d41961597222cb9e625547434e6a247277c31ec053ab001",
                "hash": "0xe696fbdff5ef9d84aae7d060ab6445b2c5a33654a950e929fcfc99699f9bef09"
            }
        ],
        "hash": "0x90d1ac4a02fce39889f4c76f894dd7ff917c02201ce074d2c04771167381a953",
	    "sig": "0x9a794b198de6b788d62aad0e3b9c5d96fde603953fbf5a79015494c43b6b634e9fcd8b5bf3dc50c91827e1b6a3d2b7f0a8693f6d61d0a56c2291004862e9600f"
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
            "parentHash": "hash_1",
            "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
            "number": 1,
            "stateRoot": "0x",
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
            "parentHash": "hash_1",
            "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
            "number": 2,
            "stateRoot": "0x01",
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
            "stateRoot": "0x",
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
            "stateRoot": "0x",
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
            "stateRoot": "0x60b239e9614da6a9c9791b0a30d0708fdb11b80283b5a2ffac1c6a5494557f5e",
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
            "stateRoot": "0x",
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
            "stateRoot": "0x",
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
            "stateRoot": "0x",
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
                "to": "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
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

var ProcessOrphanBlockData = []string{
	`{
        "header": {
            "parentHash": null,
            "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
            "number": 1,
            "stateRoot": "0x68656C6C6F",
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
            "stateRoot": "0x68656C6C6F",
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
            "stateRoot": "0xe5cd4e4e56a3244d8ed9d36cf7fc3cc91f456ab9b7208127bad8a97d2b66f5dc",
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
            "parentHash": "hash_3",
            "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
            "number": 4,
            "stateRoot": "0xcc47c5896cc9540809adf07fe42aa4b16aef4d84bc2938dc7230ff9e9363716c",
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
}
