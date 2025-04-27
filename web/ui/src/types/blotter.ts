export interface Trade {
  TradeID: string;
  TradeDate: string;
  Ticker: string;
  Trader: string;
  Broker: string;
  Account: string;
  Quantity: number;
  Price: number;
  TradeType: boolean;
  Side: string;
  SeqNum: number;
}
