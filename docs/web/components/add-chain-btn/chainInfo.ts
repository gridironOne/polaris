import type { ChainInfo, BIP44 } from '@keplr-wallet/types';
import { Bech32Address } from '@keplr-wallet/did:fury:';

const coinType60 = 60

const furyBip44: BIP44 = {
  coinType: coinType60,
};

const FURY = {
  coinDenom: 'fury',
  coinMinimalDenom: 'afury',
  coinDecimals: 18,
};

const LOCAL_CHAIN_INFO: ChainInfo = {
  coinType: coinType60,
  rpc: 'http://localhost:26657',
  rest: 'http://localhost:1317',
  chainId: 'clockend-4200',
  chainName: 'polaris',
  stakeCurrency: FURY,
  bip44: furyBip44,
  bech32Config: Bech32Address.defaultBech32Config('did:fury:'),
  currencies: [FURY],
  feeCurrencies: [{
    ...FURY,
    gasPriceStep: {
      low: 10000000000,
      average: 25000000000,
      high: 40000000000,
    },
  }],
  features: ['ibc-transfer', 'ibc-go', 'eth-address-gen', 'eth-key-sign'],
};

export default LOCAL_CHAIN_INFO;