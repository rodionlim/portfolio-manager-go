export interface Trade {
  TradeID: string;
  TradeDate: string;
  Ticker: string;
  Book: string;
  Broker: string;
  Account: string;
  Quantity: number;
  Price: number;
  Fx: number;
  TradeType: boolean;
  Side: string;
  SeqNum: number;
}
