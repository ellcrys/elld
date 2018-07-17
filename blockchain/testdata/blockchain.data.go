package testdata

var BlockchainDotGoJSON = []string{
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
				"from": "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad",
				"value": "100.333",
				"timestamp": 1508673895,
				"fee": "0.00003",
				"invokeArgs": null,
				"sig": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs",
				"hash": "93udndte7hxbvhivmnzbzguruhcbybcdbxcbyulmxsncs"
			}
		],
		"hash": "genesis_hash",
		"sig": "abc"
	}`,
	`{
		"header": {
			"parentHash": "some_hash_1",
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
		"hash": "hash_2",
		"sig": "abc"
	}`,
	`{
		"header": {
			"parentHash": "hash_2",
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
		"hash": "hash_3",
		"sig": "abc"
	}`,
}
