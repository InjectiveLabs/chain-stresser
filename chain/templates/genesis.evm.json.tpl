{
  "genesis_time": "{{.GenesisTime}}",
  "chain_id": "{{.ChainID}}",
  "initial_height": "1",
  "consensus_params": {
    "block": {
      "max_bytes": "4194304",
      "max_gas": "10000000000"
    },
    "evidence": {
      "max_age_num_blocks": "100000",
      "max_age_duration": "172800000000000",
      "max_bytes": "1048576"
    },
    "validator": {
      "pub_key_types": [
        "ed25519"
      ]
    },
    "version": {
      "app": "0"
    },
    "synchrony": {
      "precision": "505000000",
      "message_delay": "15000000000"
    },
    "feature": {
      "vote_extensions_enable_height": "0",
      "pbts_enable_height": "0"
    }
  },
  "app_hash": "",
  "app_state": {
    "07-tendermint": null,
    "auction": {
      "params": {
        "auction_period": "604800",
        "min_next_bid_increment_rate": "0.002500000000000000"
      },
      "auction_round": "0",
      "highest_bid": null,
      "auction_ending_timestamp": "0",
      "last_auction_result": null
    },
    "auth": {
      "params": {
        "max_memo_characters": "256",
        "tx_sig_limit": "7",
        "tx_size_cost_per_byte": "10",
        "sig_verify_cost_ed25519": "590",
        "sig_verify_cost_secp256k1": "1000"
      },
      "accounts": []
    },
    "authz": {
      "authorization": []
    },
    "bank": {
      "params": {
        "send_enabled": [],
        "default_send_enabled": true
      },
      "balances": [],
      "supply": [],
      "denom_metadata": [],
      "send_enabled": []
    },
    "capability": {
      "index": "1",
      "owners": []
    },
    "chainlink": {
      "params": {
        "link_denom": "peggy0x514910771AF9Ca656af840dff83E8264EcF986CA",
        "payout_block_interval": "100000",
        "module_admin": ""
      },
      "feed_configs": [],
      "latest_epoch_and_rounds": [],
      "feed_transmissions": [],
      "latest_aggregator_round_ids": [],
      "reward_pools": [],
      "feed_observation_counts": [],
      "feed_transmission_counts": [],
      "pending_payeeships": []
    },
    "consensus": null,
    "crisis": {
      "constant_fee": {
        "denom": "{{.BondDenom}}",
        "amount": "1000"
      }
    },
    "distribution": {
      "params": {
        "community_tax": "0.020000000000000000",
        "base_proposer_reward": "0.000000000000000000",
        "bonus_proposer_reward": "0.000000000000000000",
        "withdraw_addr_enabled": true
      },
      "fee_pool": {
        "community_pool": []
      },
      "delegator_withdraw_infos": [],
      "previous_proposer": "",
      "outstanding_rewards": [],
      "validator_accumulated_commissions": [],
      "validator_historical_rewards": [],
      "validator_current_rewards": [],
      "delegator_starting_infos": [],
      "validator_slash_events": []
    },
    "evidence": {
      "evidence": []
    },
    "evm": {
      "accounts": [],
      "params": {
        "evm_denom": "{{.BondDenom}}",
        "enable_create": true,
        "enable_call": true,
        "extra_eips": [],
        "chain_config": {
          "eip155_chain_id": "{{.EthChainID}}",
          "homestead_block": "0",
          "dao_fork_block": "0",
          "dao_fork_support": true,
          "eip150_block": "0",
          "eip150_hash": "0x0000000000000000000000000000000000000000000000000000000000000000",
          "eip155_block": "0",
          "eip158_block": "0",
          "byzantium_block": "0",
          "constantinople_block": "0",
          "petersburg_block": "0",
          "istanbul_block": "0",
          "muir_glacier_block": "0",
          "berlin_block": "0",
          "london_block": "0",
          "arrow_glacier_block": "0",
          "gray_glacier_block": "0",
          "merge_netsplit_block": "0",
          "shanghai_time": "0",
          "cancun_time": "0",
          "prague_time": "0",
          "blob_schedule_config": {
            "cancun": {
              "target": "3",
              "max": "6",
              "base_fee_update_fraction": "3338477"
            },
            "prague": {
              "target": "6",
              "max": "9",
              "base_fee_update_fraction": "5007716"
            },
            "osaka": {
              "target": "6",
              "max": "9",
              "base_fee_update_fraction": "5007716"
            },
            "verkle": null
          }
        },
        "allow_unprotected_txs": true
      }
    },
    "exchange": {
      "params": {
        "spot_market_instant_listing_fee": {
          "denom": "{{.BondDenom}}",
          "amount": "20000000000000000000"
        },
        "derivative_market_instant_listing_fee": {
          "denom": "{{.BondDenom}}",
          "amount": "20000000000000000000"
        },
        "default_spot_maker_fee_rate": "-0.000100000000000000",
        "default_spot_taker_fee_rate": "0.001000000000000000",
        "default_derivative_maker_fee_rate": "-0.000100000000000000",
        "default_derivative_taker_fee_rate": "0.001000000000000000",
        "default_initial_margin_ratio": "0.050000000000000000",
        "default_maintenance_margin_ratio": "0.020000000000000000",
        "default_funding_interval": "3600",
        "funding_multiple": "3600",
        "relayer_fee_share_rate": "0.400000000000000000",
        "default_hourly_funding_rate_cap": "0.000625000000000000",
        "default_hourly_interest_rate": "0.000004166660000000",
        "max_derivative_order_side_count": 20,
        "inj_reward_staked_requirement_threshold": "100000000000000000000",
        "trading_rewards_vesting_duration": "604800",
        "liquidator_reward_share_rate": "0.050000000000000000",
        "binary_options_market_instant_listing_fee": {
          "denom": "{{.BondDenom}}",
          "amount": "100000000000000000000"
        },
        "atomic_market_order_access_level": "SmartContractsOnly",
        "spot_atomic_market_order_fee_multiplier": "2.500000000000000000",
        "derivative_atomic_market_order_fee_multiplier": "2.500000000000000000",
        "binary_options_atomic_market_order_fee_multiplier": "2.500000000000000000",
        "minimal_protocol_fee_rate": "0.000050000000000000",
        "is_instant_derivative_market_launch_enabled": false,
        "post_only_mode_height_threshold": "0",
        "margin_decrease_price_timestamp_threshold_seconds": "0",
        "exchange_admins": [],
        "inj_auction_max_cap": "0"
      },
      "spot_markets": [],
      "derivative_markets": [],
      "spot_orderbook": [],
      "derivative_orderbook": [],
      "balances": [],
      "positions": [],
      "subaccount_trade_nonces": [],
      "expiry_futures_market_info_state": [],
      "perpetual_market_info": [],
      "perpetual_market_funding_state": [],
      "derivative_market_settlement_scheduled": [],
      "is_spot_exchange_enabled": true,
      "is_derivatives_exchange_enabled": true,
      "trading_reward_campaign_info": null,
      "trading_reward_pool_campaign_schedule": [],
      "trading_reward_campaign_account_points": [],
      "fee_discount_schedule": null,
      "fee_discount_account_tier_ttl": [],
      "fee_discount_bucket_volume_accounts": [],
      "is_first_fee_cycle_finished": false,
      "pending_trading_reward_pool_campaign_schedule": [],
      "pending_trading_reward_campaign_account_points": [],
      "rewards_opt_out_addresses": [],
      "historical_trade_records": [],
      "binary_options_markets": [],
      "binary_options_market_ids_scheduled_for_settlement": [],
      "spot_market_ids_scheduled_to_force_close": [],
      "denom_decimals": [],
      "conditional_derivative_orderbooks": [],
      "market_fee_multipliers": [],
      "orderbook_sequences": [],
      "subaccount_volumes": [],
      "market_volumes": [],
      "grant_authorizations": [],
      "active_grants": []
    },
    "feegrant": {
      "allowances": []
    },
    "feeibc": {
      "identified_fees": [],
      "fee_enabled_channels": [],
      "registered_payees": [],
      "registered_counterparty_payees": [],
      "forward_relayers": []
    },
    "feemarket": {
      "params": {
        "no_base_fee": true,
        "base_fee_change_denominator": 8,
        "elasticity_multiplier": 2,
        "enable_height": "0",
        "base_fee": "1000000000",
        "min_gas_price": "0.000000000000000000",
        "min_gas_multiplier": "0.500000000000000000"
      },
      "block_gas": "10000000000"
    },
    "genutil": {
      "gen_txs": []
    },
    "gov": {
      "starting_proposal_id": "1",
      "deposits": [],
      "votes": [],
      "proposals": [],
      "deposit_params": null,
      "voting_params": null,
      "tally_params": null,
      "params": {
        "min_deposit": [
          {
            "denom": "{{.BondDenom}}",
            "amount": "10000000"
          }
        ],
        "max_deposit_period": "172800s",
        "voting_period": "172800s",
        "quorum": "0.334000000000000000",
        "threshold": "0.500000000000000000",
        "veto_threshold": "0.334000000000000000",
        "min_initial_deposit_ratio": "0.000000000000000000",
        "proposal_cancel_ratio": "0.500000000000000000",
        "proposal_cancel_dest": "",
        "expedited_voting_period": "86400s",
        "expedited_threshold": "0.667000000000000000",
        "expedited_min_deposit": [
          {
            "denom": "{{.BondDenom}}",
            "amount": "50000000"
          }
        ],
        "burn_vote_quorum": false,
        "burn_proposal_deposit_prevote": false,
        "burn_vote_veto": true,
        "min_deposit_ratio": "0.010000000000000000"
      },
      "constitution": ""
    },
    "ibc": {
      "client_genesis": {
        "clients": [],
        "clients_consensus": [],
        "clients_metadata": [],
        "params": {
          "allowed_clients": [
            "*"
          ]
        },
        "create_localhost": false,
        "next_client_sequence": "0"
      },
      "connection_genesis": {
        "connections": [],
        "client_connection_paths": [],
        "next_connection_sequence": "0",
        "params": {
          "max_expected_time_per_block": "30000000000"
        }
      },
      "channel_genesis": {
        "channels": [],
        "acknowledgements": [],
        "commitments": [],
        "receipts": [],
        "send_sequences": [],
        "recv_sequences": [],
        "ack_sequences": [],
        "next_channel_sequence": "0",
        "params": {
          "upgrade_timeout": {
            "height": {
              "revision_number": "0",
              "revision_height": "0"
            },
            "timestamp": "600000000000"
          }
        }
      }
    },
    "ibchooks": null,
    "insurance": {
      "params": {
        "default_redemption_notice_period_duration": "1209600s"
      },
      "insurance_funds": [],
      "redemption_schedule": [],
      "next_share_denom_id": "1",
      "next_redemption_schedule_id": "1"
    },
    "interchainaccounts": {
      "controller_genesis_state": {
        "active_channels": [],
        "interchain_accounts": [],
        "ports": [],
        "params": {
          "controller_enabled": true
        }
      },
      "host_genesis_state": {
        "active_channels": [],
        "interchain_accounts": [],
        "port": "icahost",
        "params": {
          "host_enabled": true,
          "allow_messages": [
            "*"
          ]
        }
      }
    },
    "mint": {
      "minter": {
        "inflation": "0.130000000000000000",
        "annual_provisions": "0.000000000000000000"
      },
      "params": {
        "mint_denom": "{{.BondDenom}}",
        "inflation_rate_change": "0.130000000000000000",
        "inflation_max": "0.200000000000000000",
        "inflation_min": "0.070000000000000000",
        "goal_bonded": "0.670000000000000000",
        "blocks_per_year": "6311520"
      }
    },
    "oracle": {
      "params": {
        "pyth_contract": ""
      },
      "band_relayers": [],
      "band_price_states": [],
      "price_feed_price_states": [],
      "coinbase_price_states": [],
      "band_ibc_price_states": [],
      "band_ibc_oracle_requests": [],
      "band_ibc_params": {
        "band_ibc_enabled": false,
        "ibc_request_interval": "7",
        "ibc_source_channel": "",
        "ibc_version": "bandchain-1",
        "ibc_port_id": "oracle",
        "legacy_oracle_ids": []
      },
      "band_ibc_latest_client_id": "0",
      "calldata_records": [],
      "band_ibc_latest_request_id": "0",
      "chainlink_price_states": [],
      "historical_price_records": [],
      "provider_states": [],
      "pyth_price_states": [],
      "stork_price_states": [],
      "stork_publishers": []
    },
    "packetfowardmiddleware": {
      "in_flight_packets": {}
    },
    "params": null,
    "peggy": {
      "params": {
        "peggy_id": "injective-peggyid",
        "contract_source_hash": "",
        "bridge_ethereum_address": "",
        "bridge_chain_id": "0",
        "signed_valsets_window": "10000",
        "signed_batches_window": "10000",
        "signed_claims_window": "10000",
        "target_batch_timeout": "43200000",
        "average_block_time": "5000",
        "average_ethereum_block_time": "15000",
        "slash_fraction_valset": "0.001000000000000000",
        "slash_fraction_batch": "0.001000000000000000",
        "slash_fraction_claim": "0.001000000000000000",
        "slash_fraction_conflicting_claim": "0.001000000000000000",
        "unbond_slashing_valsets_window": "10000",
        "slash_fraction_bad_eth_signature": "0.001000000000000000",
        "cosmos_coin_denom": "{{.BondDenom}}",
        "cosmos_coin_erc20_contract": "",
        "claim_slashing_enabled": false,
        "bridge_contract_start_height": "0",
        "valset_reward": {
          "denom": "",
          "amount": "0"
        },
        "admins": []
      },
      "last_observed_nonce": "0",
      "valsets": [],
      "valset_confirms": [],
      "batches": [],
      "batch_confirms": [],
      "attestations": [],
      "orchestrator_addresses": [],
      "erc20_to_denoms": [],
      "unbatched_transfers": [],
      "last_observed_ethereum_height": "0",
      "last_outgoing_batch_id": "0",
      "last_outgoing_pool_id": "0",
      "last_observed_valset": {
        "nonce": "0",
        "members": [],
        "height": "0",
        "reward_amount": "0",
        "reward_token": ""
      },
      "ethereum_blacklist": []
    },
    "permissions": {
      "params": {
        "wasm_hook_query_max_gas": "200000"
      },
      "namespaces": []
    },
    "slashing": {
      "params": {
        "signed_blocks_window": "100",
        "min_signed_per_window": "0.500000000000000000",
        "downtime_jail_duration": "600s",
        "slash_fraction_double_sign": "0.050000000000000000",
        "slash_fraction_downtime": "0.010000000000000000"
      },
      "signing_infos": [],
      "missed_blocks": []
    },
    "staking": {
      "params": {
        "unbonding_time": "1814400s",
        "max_validators": 100,
        "max_entries": 7,
        "historical_entries": 10000,
        "bond_denom": "{{.BondDenom}}",
        "min_commission_rate": "0.000000000000000000"
      },
      "last_total_power": "0",
      "last_validator_powers": [],
      "validators": [],
      "delegations": [],
      "unbonding_delegations": [],
      "redelegations": [],
      "exported": false
    },
    "tokenfactory": {
      "params": {
        "denom_creation_fee": [
          {
            "denom": "{{.BondDenom}}",
            "amount": "10000000000000000000"
          }
        ]
      },
      "factory_denoms": []
    },
    "transfer": {
      "port_id": "transfer",
      "denom_traces": [],
      "params": {
        "send_enabled": true,
        "receive_enabled": true
      },
      "total_escrowed": []
    },
    "upgrade": {},
    "wasm": {
      "params": {
        "code_upload_access": {
          "permission": "Everybody",
          "addresses": []
        },
        "instantiate_default_permission": "Everybody"
      },
      "codes": [],
      "contracts": [],
      "sequences": []
    },
    "xwasm": {
      "params": {
        "is_execution_enabled": true,
        "max_begin_block_total_gas": "42000000",
        "max_contract_gas_limit": "3500000",
        "min_gas_price": "1000000000",
        "register_contract_access": {
          "permission": "Unspecified",
          "addresses": []
        }
      },
      "registered_contracts": []
    }
  }
}