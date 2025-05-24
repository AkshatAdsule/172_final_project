import { createFileRoute, Outlet } from "@tanstack/react-router";
import "./rides.css";
import Sidebar from "../components/sidebar";

export const Route = createFileRoute("/rides")({
	component: RouteComponent,
});

function RouteComponent() {
	return (
		<div className="container">
			<Sidebar />
			<main>
				<Outlet />
			</main>
		</div>
	);
}
