import type { ReactNode } from "react";
import { LatLngContext } from "../contexts/LatLngContext";
import { useLatLngList } from "../hooks/useLatLngList";

export function LatLngProvider({ children }: { children: ReactNode }) {
	const { readyState, sendMessage, currentLocation, rideState } =
		useLatLngList();

	return (
		<LatLngContext.Provider
			value={{
				readyState,
				sendMessage,
				currentLocation,
				rideState,
			}}
		>
			{children}
		</LatLngContext.Provider>
	);
}
export { LatLngContext };
