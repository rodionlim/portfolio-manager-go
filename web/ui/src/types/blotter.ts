export interface Trade {
  TradeID: string;
  TradeDate: string;
  Ticker: string;
  InstrumentType?: string;
  UnderlyingTicker?: string;
  UnderlyingSpotRef?: number;
  ExpiryDate?: string;
  StrikePrice?: number;
  CallPut?: string;
  Book: string;
  Broker: string;
  Account: string;
  Quantity: number;
  Price: number;
  Fx: number;
  Status?: string;
  OrigTradeID?: string;
  TradeType: boolean;
  Side: string;
  SeqNum: number;
}
