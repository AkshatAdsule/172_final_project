import { createFileRoute, Outlet } from "@tanstack/react-router";
import "./rides.css";
import Sidebar from "../components/sidebar";
import { LatLngProvider } from "../providers/LatLngProvider";
import { APIProvider } from "@vis.gl/react-google-maps";

export const Route = createFileRoute("/rides")({
	component: RouteComponent,
});

function RouteComponent() {
	return (
		<div className="container">
			<Sidebar />
			<main>
				<APIProvider apiKey={import.meta.env.VITE_GOOGLE_MAPS_API_KEY}>
					<LatLngProvider url="ws://localhost:8080/">
						<Outlet />
					</LatLngProvider>
				</APIProvider>
			</main>
		</div>
	);
}
