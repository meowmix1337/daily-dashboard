export interface HourlyForecast {
  time: string;
  temp: number;
  icon: string;
}

export interface WeatherData {
  temp: number;
  high: number;
  low: number;
  condition: string;
  icon: string;
  humidity: number;
  windSpeed: number;
  uvIndex: number;
  aqi: number;
  aqiLabel: string;
  hourly: HourlyForecast[];
}

export interface CalendarEvent {
  time: string;
  title: string;
  color: string;
  duration: string;
}

export interface Task {
  id: string;
  text: string;
  done: boolean;
  priority: 'high' | 'medium' | 'low';
}

export interface NewsItem {
  title: string;
  source: string;
  time: string;
  url: string;
}

export interface NewsCategory {
  name: string;
  items: NewsItem[];
}

export interface StockQuote {
  symbol: string;
  price: number;
  change: number;
  pct: number;
}

export interface SymbolSearchResult {
  symbol: string;
  description: string;
  type: string;
}

export interface Quote {
  text: string;
  author: string;
}

export interface MetaData {
  sunrise: string;
  sunset: string;
  daylight: string;
  quote: Quote;
}

export interface DashboardResponse {
  weather: WeatherData | null;
  calendar: CalendarEvent[];
  tasks: Task[];
  tasksTotal?: number;
  stocks: StockQuote[] | null;
  meta: MetaData | null;
}

export interface UserSettings {
  latitude: number | null;
  longitude: number | null;
  calendar_ics_url: string | null;
  timezone: string | null;
}

export interface NewsCategoryType {
  id: string;
  label: string;
  sort_order: number;
}

export interface NewsCategoriesResponse {
  available: NewsCategoryType[];
  selected: NewsCategoryType[];
}

export interface TaskLabel {
  id: string;
  name: string;
  color: string;
  created_at: string;
}
