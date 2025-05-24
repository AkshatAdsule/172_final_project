import { createContext } from "react";
import type { UseWebSocketResult } from "../hooks/websocket";

export const WSContext = createContext<UseWebSocketResult | null>(null);
