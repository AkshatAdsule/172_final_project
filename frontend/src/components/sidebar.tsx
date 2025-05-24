import { Link } from "@tanstack/react-router";
import "./sidebar.css";

export default function Sidebar() {
	const mockRideHistory = [
		{ id: 1, name: "Morning Ride", date: "2023-10-01" },
		{ id: 2, name: "Evening Ride", date: "2023-10-02" },
		{ id: 3, name: "Weekend Ride", date: "2023-10-03" },
	];

	return (
		<nav>
			<div className="brand">
				<h1>B3</h1>
			</div>
			<hr />
			<div className="rides">
				<Link to="/rides/live">
					<div>Live Track</div>
				</Link>
				<hr />
				{mockRideHistory.map((ride) => (
					<Link key={ride.id} to={`/rides`}>
						<div className="ride">
							<span className="name">{ride.name}</span>
							<span className="date">
								{new Date(ride.date).toLocaleDateString()}
							</span>
						</div>
					</Link>
				))}
			</div>
		</nav>
	);
}
