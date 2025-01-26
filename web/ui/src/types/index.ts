export interface ReferenceDataItem {
  id: string;
  name: string;
  underlying_ticker: string;
  yahoo_ticker: string;
  google_ticker: string;
  dividends_sg_ticker: string;
  asset_class: string;
  asset_sub_class: string;
  category: string;
  sub_category: string;
  ccy: string;
  domicile: string;
  coupon_rate: number;
  maturity_date: string;
  strike_price: number;
  call_put: string;
}

export type ReferenceData = Record<string, ReferenceDataItem>;
