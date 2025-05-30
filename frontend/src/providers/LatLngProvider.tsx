import type { ReactNode } from "react";
import { LatLngContext } from "../contexts/LatLngContext";
import { useLatLngList } from "../hooks/useLatLngList";

export function LatLngProvider({
	children,
	url,
}: { children: ReactNode; url: string }) {
	const { latLngList, readyState, sendMessage } = useLatLngList(url);

	return (
		<LatLngContext.Provider value={{ latLngList, readyState, sendMessage }}>
			{children}
		</LatLngContext.Provider>
	);
}
export { LatLngContext };
