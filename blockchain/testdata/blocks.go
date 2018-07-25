package testdata

var TestGenesisBlock = `{
    "header": {
        "parentHash": null,
        "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
        "number": 1,
        "stateRoot": "0xd66d3a287ca3730d2eeae550f765627c767a6b809cf65db1221879b532534297",
        "transactionsRoot": "0x",
        "nonce": 0,
        "mixHash": "",
        "difficulty": "500000",
        "timestamp": 0
    },
    "transactions": [
        {
            "type": 1,
            "nonce": 2,
            "to": "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
            "senderPubKey": "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
            "from": "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
            "value": "1",
            "timestamp": 1508673895,
            "fee": "0.1",
            "invokeArgs": null,
            "sig": "0xe9821f4d75acf65cab8ba7df82b161c2cc5fb5acb56cb4c54d8bb0425131106489d1041ed964875d7416cd25baf792a2068dc27e07b91110b3e8469f903a420b",
            "hash": "0xc3ca4c0910a49009a51af2d43d01d5c5cead0f1abd9529301c82dec871383886"
        }
    ],
    "hash": "hash_1",
    "sig": "abc"
}`

var TestBlocks = []string{
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
			{
				"type": 1,
				"nonce": 2,
				"to": "efjshfhh389djn29snmnvuis",
				"senderPubKey": "xsj2909jfhhjskmj99k",
				"from": "",
				"value": "100.333",
				"timestamp": 1508673895,
				"fee": "0.00003",
				"invokeArgs": null,
				"sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
				"hash": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs"
			},
			{
				"type": 2,
				"nonce": 2,
				"to": null,
				"senderPubKey": "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
				"from": "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				"value": "10",
				"timestamp": 1508673895,
				"fee": "0.00003",
				"invokeArgs": null,
				"sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
				"hash": "shhfd7387ydhudhsy8ehhhfjsg748hd"
			}
		],
		"hash": "lkssjajsdnaskomcskosks",
		"sig": "abc"
	}`,
	`{
		"header": {
			"parentHash": null,
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
			{
				"type": 1,
				"nonce": 2,
				"to": "efjshfhh389djn29snmnvuis",
				"senderPubKey": "xsj2909jfhhjskmj99k",
				"from": "",
				"value": "100.333",
				"timestamp": 1508673895,
				"fee": "0.00003",
				"invokeArgs": null,
				"sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
				"hash": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs"
			},
			{
				"type": 2,
				"nonce": 2,
				"to": null,
				"senderPubKey": "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
				"from": "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				"value": "10",
				"timestamp": 1508673895,
				"fee": "0.00003",
				"invokeArgs": null,
				"sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
				"hash": "shhfd7387ydhudhsy8ehhhfjsg748hd"
			}
		],
		"hash": "lkssjajsdnaskomcskosks",
		"sig": "abc"
	}`,
	`{
		"header": {
			"parentHash": null,
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
			{
				"type": 1,
				"nonce": 2,
				"to": "efjshfhh389djn29snmnvuis",
				"senderPubKey": "xsj2909jfhhjskmj99k",
				"from": "",
				"value": "100.333",
				"timestamp": 1508673895,
				"fee": "0.00003",
				"invokeArgs": null,
				"sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
				"hash": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs"
			},
			{
				"type": 2,
				"nonce": 2,
				"to": null,
				"senderPubKey": "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
				"from": "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				"value": "10",
				"timestamp": 1508673895,
				"fee": "0.00003",
				"invokeArgs": null,
				"sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
				"hash": "shhfd7387ydhudhsy8ehhhfjsg748hd"
			}
		],
		"hash": "lkssjajsdnaskomcskosks",
		"sig": "abc"
	}`,
}

var TestBlock1 = `{
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
        {
            "type": 1,
            "nonce": 2,
            "to": "efjshfhh389djn29snmnvuis",
            "senderPubKey": "xsj2909jfhhjskmj99k",
            "from": "",
            "value": "100.333",
            "timestamp": 1508673895,
            "fee": "0.00003",
            "invokeArgs": null,
            "sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
            "hash": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs"
        },
        {
            "type": 2,
            "nonce": 2,
            "to": null,
            "senderPubKey": "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
            "from": "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
            "value": "10",
            "timestamp": 1508673895,
            "fee": "0.00003",
            "invokeArgs": null,
            "sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
            "hash": "shhfd7387ydhudhsy8ehhhfjsg748hd"
        }
    ],
    "hash": "lkssjajsdnaskomcskosks",
    "sig": "abc"
}`

var TestBlock2 = `{
    "header": {
        "parentHash": null,
        "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
        "number": 2,
        "stateRoot": "abc",
        "transactionsRoot": "sjifhahe3i9nfi",
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
            "from": "",
            "value": "100.333",
            "timestamp": 1508673895,
            "fee": "0.00003",
            "invokeArgs": null,
            "sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
            "hash": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs"
        },
        {
            "type": 2,
            "nonce": 2,
            "to": null,
            "senderPubKey": "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
            "from": "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
            "value": "10",
            "timestamp": 1508673895,
            "fee": "0.00003",
            "invokeArgs": null,
            "sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
            "hash": "shhfd7387ydhudhsy8ehhhfjsg748hd"
        }
    ],
    "hash": "lkssjajsdnaskomcskosks",
    "sig": "abc"
}`

var TestBlock3 = `{
    "header": {
        "parentHash": null,
        "creatorPubKey": "49VzsGezoNjJQxHjekoCQP9CXZUs34CmCY53kGaHyR9rCJQJbJW",
        "number": 3,
        "stateRoot": "abc",
        "transactionsRoot": "sjifhahe3dasasdai9nfi",
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
            "from": "",
            "value": "100.333",
            "timestamp": 1508673895,
            "fee": "0.00003",
            "invokeArgs": null,
            "sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
            "hash": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs"
        },
        {
            "type": 2,
            "nonce": 2,
            "to": null,
            "senderPubKey": "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
            "from": "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
            "value": "10",
            "timestamp": 1508673895,
            "fee": "0.00003",
            "invokeArgs": null,
            "sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
            "hash": "shhfd7387ydhudhsy8ehhhfjsg748hd"
        }
    ],
    "hash": "lkssjajsdnaskomcskosks",
    "sig": "abc"
}`
