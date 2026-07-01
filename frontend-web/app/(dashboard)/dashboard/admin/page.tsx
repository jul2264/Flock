"use client";

import { useEffect, useState } from "react";
import { useAuth } from "@clerk/nextjs";
import { api, User, Event, PlatformStats, UsersListResponse } from "../../../lib/api";
import { 
	ShieldAlert, 
	Users, 
	Calendar, 
	Award, 
	TrendingUp, 
	Activity,
	CheckCircle,
	AlertTriangle,
	Trash2,
	ChevronLeft,
	ChevronRight,
	Search,
	MapPin
} from "lucide-react";

export default function AdminPage() {
	const { getToken, isSignedIn } = useAuth();
	const [stats, setStats] = useState<PlatformStats | null>(null);
	const [usersRes, setUsersRes] = useState<UsersListResponse | null>(null);
	const [events, setEvents] = useState<Event[]>([]);
	const [loading, setLoading] = useState(true);
	const [searchQuery, setSearchQuery] = useState("");
	const [page, setPage] = useState(1);
	const [successMsg, setSuccessMsg] = useState("");
	const [errorMsg, setErrorMsg] = useState("");

	useEffect(() => {
		async function loadAdminData() {
			try {
				const token = await getToken();
				if (token) {
					const [statsData, usersListData, eventsListData] = await Promise.all([
						api.getAdminStats(token),
						api.listAdminUsers(token, page, 8),
						api.listEvents(token),
					]);
					setStats(statsData);
					setUsersRes(usersListData);
					setEvents(eventsListData);
				}
			} catch (err) {
				console.error("Failed to load admin panel data:", err);
				setErrorMsg("Failed to retrieve administrative details.");
			} finally {
				setLoading(false);
			}
		}

		if (isSignedIn) {
			loadAdminData();
		}
	}, [isSignedIn, getToken, page]);

	// Fetch paginated user lists when page changes
	const handlePageChange = async (newPage: number) => {
		if (newPage < 1 || (usersRes && newPage > Math.ceil(usersRes.total / usersRes.limit))) return;
		setPage(newPage);
	};

	const handleRoleUpdate = async (userId: string, newRole: string) => {
		setErrorMsg("");
		setSuccessMsg("");
		try {
			const token = await getToken();
			if (!token) return;

			const updatedUser = await api.updateUserRole(token, userId, newRole);
			
			// Update user list state locally
			if (usersRes) {
				const updatedUsersList = usersRes.users.map(u => u.id === userId ? updatedUser : u);
				setUsersRes({ ...usersRes, users: updatedUsersList });
			}

			// Reload stats
			const statsData = await api.getAdminStats(token);
			setStats(statsData);

			setSuccessMsg(`User role successfully updated to "${newRole}".`);
		} catch (err: unknown) {
			console.error("Role update error:", err);
			const error = err as Error;
			setErrorMsg(error.message || "Failed to update user role.");
		}
	};

	const handleModerateDeleteEvent = async (eventId: string, eventTitle: string) => {
		if (!confirm(`Are you sure you want to moderate and cancel the event "${eventTitle}"?`)) return;

		setErrorMsg("");
		setSuccessMsg("");
		try {
			const token = await getToken();
			if (!token) return;

			await api.moderateDeleteEvent(token, eventId);

			// Update events state locally (mark as cancelled)
			const updatedEvents = events.map(e => e.id === eventId ? { ...e, status: "cancelled" } : e);
			setEvents(updatedEvents);

			// Reload stats
			const statsData = await api.getAdminStats(token);
			setStats(statsData);

			setSuccessMsg(`Event "${eventTitle}" has been cancelled/moderated successfully.`);
		} catch (err: unknown) {
			console.error("Moderation error:", err);
			const error = err as Error;
			setErrorMsg(error.message || "Failed to moderate event.");
		}
	};

	// Client-side search filters
	const filteredUsers = usersRes?.users.filter(u => 
		u.full_name?.toLowerCase().includes(searchQuery.toLowerCase()) || 
		u.email.toLowerCase().includes(searchQuery.toLowerCase()) ||
		u.username?.toLowerCase().includes(searchQuery.toLowerCase())
	) || [];

	if (loading) {
		return (
			<div className="flex flex-col items-center justify-center min-h-[50vh] gap-4">
				<div className="h-8 w-8 border-4 border-red-600 border-t-transparent rounded-full animate-spin" />
				<p className="text-zinc-500 text-sm">Loading admin metrics & telemetry...</p>
			</div>
		);
	}

	return (
		<div className="space-y-8">
			{/* Admin Header */}
			<div className="text-left">
				<h1 className="text-3xl font-extrabold tracking-tight">Admin Console</h1>
				<p className="text-zinc-400 text-sm mt-1">Review system analytics, manage member access roles, and moderate content.</p>
			</div>

			{/* Success / Error Alerts */}
			{successMsg && (
				<div className="p-4 rounded-xl border border-emerald-500/10 bg-emerald-500/5 text-emerald-400 text-sm flex items-center gap-2">
					<CheckCircle className="h-4 w-4 flex-shrink-0" />
					<span>{successMsg}</span>
				</div>
			)}
			{errorMsg && (
				<div className="p-4 rounded-xl border border-red-500/10 bg-red-500/5 text-red-400 text-sm flex items-center gap-2">
					<AlertTriangle className="h-4 w-4 flex-shrink-0" />
					<span>{errorMsg}</span>
				</div>
			)}

			{/* Platforms Stats Cards */}
			{stats && (
				<div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
					<div className="p-5 rounded-2xl border border-zinc-900 bg-zinc-950/40 backdrop-blur-md flex items-center gap-4 text-left">
						<div className="h-12 w-12 rounded-xl bg-violet-600/10 border border-violet-500/20 flex items-center justify-center text-violet-400">
							<Users className="h-6 w-6" />
						</div>
						<div>
							<p className="text-xs text-zinc-500 font-medium">Total Registered Users</p>
							<p className="text-2xl font-extrabold text-white mt-1">{stats.total_users}</p>
						</div>
					</div>

					<div className="p-5 rounded-2xl border border-zinc-900 bg-zinc-950/40 backdrop-blur-md flex items-center gap-4 text-left">
						<div className="h-12 w-12 rounded-xl bg-red-600/10 border border-red-500/20 flex items-center justify-center text-red-400">
							<Activity className="h-6 w-6" />
						</div>
						<div>
							<p className="text-xs text-zinc-500 font-medium">Active (Last 24h)</p>
							<p className="text-2xl font-extrabold text-white mt-1">{stats.active_users_24h}</p>
						</div>
					</div>

					<div className="p-5 rounded-2xl border border-zinc-900 bg-zinc-950/40 backdrop-blur-md flex items-center gap-4 text-left">
						<div className="h-12 w-12 rounded-xl bg-indigo-600/10 border border-indigo-500/20 flex items-center justify-center text-indigo-400">
							<Calendar className="h-6 w-6" />
						</div>
						<div>
							<p className="text-xs text-zinc-500 font-medium">Scheduled Events</p>
							<p className="text-2xl font-extrabold text-white mt-1">{stats.total_events}</p>
						</div>
					</div>

					<div className="p-5 rounded-2xl border border-zinc-900 bg-zinc-950/40 backdrop-blur-md flex items-center gap-4 text-left">
						<div className="h-12 w-12 rounded-xl bg-amber-600/10 border border-amber-500/20 flex items-center justify-center text-amber-400">
							<Award className="h-6 w-6" />
						</div>
						<div>
							<p className="text-xs text-zinc-500 font-medium">Total RSVPs Confirmed</p>
							<p className="text-2xl font-extrabold text-white mt-1">{stats.total_rsvps}</p>
						</div>
					</div>
				</div>
			)}

			{/* Comparative Analytics SVG Charts */}
			{stats && (
				<div className="grid grid-cols-1 md:grid-cols-2 gap-6 text-left">
					{/* User Roles distribution */}
					<div className="p-6 rounded-2xl border border-zinc-900 bg-zinc-950/30 backdrop-blur-md space-y-4">
						<h3 className="font-bold text-zinc-200">User Role Distribution</h3>
						<div className="flex items-center gap-4 pt-4">
							{/* Simple custom SVG comparative bar chart */}
							<div className="flex-1 space-y-3">
								{Object.entries(stats.users_by_role).map(([role, count]) => {
									const percent = (count / stats.total_users) * 100;
									return (
										<div key={role} className="space-y-1.5">
											<div className="flex justify-between text-xs font-medium">
												<span className="capitalize text-zinc-300">{role}s</span>
												<span className="text-zinc-500">{count} ({Math.round(percent)}%)</span>
											</div>
											<div className="h-2 w-full bg-zinc-900 rounded-full overflow-hidden">
												<div 
													style={{ width: `${percent}%` }}
													className={`h-full rounded-full bg-gradient-to-r ${
														role === "admin" 
															? "from-red-600 to-red-500" 
															: role === "organizer" 
															? "from-violet-600 to-violet-500" 
															: "from-zinc-650 to-zinc-500"
													}`}
												/>
											</div>
										</div>
									);
								})}
							</div>
						</div>
					</div>

					{/* Event Status Distribution */}
					<div className="p-6 rounded-2xl border border-zinc-900 bg-zinc-950/30 backdrop-blur-md space-y-4">
						<h3 className="font-bold text-zinc-200">Event Status Telemetry</h3>
						<div className="flex items-center gap-4 pt-4">
							<div className="flex-1 space-y-3">
								{Object.entries(stats.events_by_status).map(([status, count]) => {
									const maxVal = Math.max(...Object.values(stats.events_by_status), 1);
									const percent = (count / maxVal) * 100;
									return (
										<div key={status} className="space-y-1.5">
											<div className="flex justify-between text-xs font-medium">
												<span className="capitalize text-zinc-300">{status}</span>
												<span className="text-zinc-500">{count}</span>
											</div>
											<div className="h-2 w-full bg-zinc-900 rounded-full overflow-hidden">
												<div 
													style={{ width: `${percent}%` }}
													className={`h-full rounded-full bg-gradient-to-r ${
														status === "published" 
															? "from-emerald-600 to-emerald-500" 
															: status === "cancelled" 
															? "from-red-600 to-red-500" 
															: "from-zinc-700 to-zinc-600"
													}`}
												/>
											</div>
										</div>
									);
								})}
							</div>
						</div>
					</div>
				</div>
			)}

			{/* User Moderation Section */}
			<div className="p-6 rounded-2xl border border-zinc-900 bg-zinc-950/30 backdrop-blur-md space-y-4 text-left">
				<div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b border-zinc-900 pb-4">
					<h2 className="text-xl font-bold text-zinc-200">User Moderation</h2>
					{/* Search user */}
					<div className="relative max-w-xs w-full">
						<Search className="absolute left-3 top-2.5 h-4.5 w-4.5 text-zinc-500" />
						<input
							type="text"
							value={searchQuery}
							onChange={(e) => setSearchQuery(e.target.value)}
							placeholder="Search by name or email..."
							className="w-full bg-zinc-900/60 border border-zinc-800 rounded-xl pl-10 pr-4 py-2 text-xs text-white focus:outline-none focus:border-violet-500"
						/>
					</div>
				</div>

				{/* Users Table */}
				<div className="overflow-x-auto">
					<table className="w-full text-sm border-collapse text-left">
						<thead>
							<tr className="border-b border-zinc-900 text-zinc-500 text-xs uppercase font-semibold">
								<th className="py-3 px-4">User Info</th>
								<th className="py-3 px-4">Role Status</th>
								<th className="py-3 px-4">Last Active</th>
								<th className="py-3 px-4 text-right">Promote / Demote</th>
							</tr>
						</thead>
						<tbody className="divide-y divide-zinc-900">
							{filteredUsers.length > 0 ? (
								filteredUsers.map((u) => (
									<tr key={u.id} className="hover:bg-zinc-900/10 transition-colors">
										<td className="py-4 px-4 flex items-center gap-3 min-w-[200px]">
											<img 
												src={u.avatar_url || "https://www.gravatar.com/avatar/00000000000000000000000000000000?d=mp&f=y"}
												alt="User avatar"
												className="h-9 w-9 rounded-full bg-zinc-800 border border-zinc-800"
											/>
											<div className="min-w-0">
												<p className="font-semibold text-zinc-200 truncate">
													{u.full_name || "Console User"}
												</p>
												<p className="text-xs text-zinc-500 truncate">{u.email}</p>
											</div>
										</td>
										<td className="py-4 px-4">
											<span className={`text-[10px] font-bold px-2 py-0.5 rounded-full uppercase ${
												u.role === "admin"
													? "bg-red-500/10 text-red-400 border border-red-500/20"
													: u.role === "organizer"
													? "bg-violet-500/10 text-violet-400 border border-violet-500/20"
													: "bg-zinc-850 text-zinc-400"
											}`}>
												{u.role}
											</span>
										</td>
										<td className="py-4 px-4 text-xs text-zinc-400">
											{u.last_seen_at ? new Date(u.last_seen_at).toLocaleDateString() : "Never"}
										</td>
										<td className="py-4 px-4 text-right">
											<select
												value={u.role}
												onChange={(e) => handleRoleUpdate(u.id, e.target.value)}
												className="bg-zinc-900 border border-zinc-800 rounded-lg px-2.5 py-1 text-xs text-zinc-200 focus:outline-none focus:border-violet-500 cursor-pointer"
											>
												<option value="user">User</option>
												<option value="organizer">Organizer</option>
												<option value="admin">Admin</option>
											</select>
										</td>
									</tr>
								))
							) : (
								<tr>
									<td colSpan={4} className="py-8 text-center text-xs text-zinc-500">
										No user profiles matched your search parameters.
									</td>
								</tr>
							)}
						</tbody>
					</table>
				</div>

				{/* Pagination Controls */}
				{usersRes && usersRes.total > usersRes.limit && (
					<div className="flex justify-between items-center border-t border-zinc-900 pt-4 text-xs text-zinc-500">
						<span>Showing page {usersRes.page} of {Math.ceil(usersRes.total / usersRes.limit)}</span>
						<div className="flex gap-2">
							<button
								onClick={() => handlePageChange(page - 1)}
								disabled={page === 1}
								className="cursor-pointer p-1.5 rounded-lg border border-zinc-800 bg-zinc-900 text-zinc-300 disabled:opacity-40 disabled:cursor-not-allowed hover:bg-zinc-800"
							>
								<ChevronLeft className="h-4 w-4" />
							</button>
							<button
								onClick={() => handlePageChange(page + 1)}
								disabled={page === Math.ceil(usersRes.total / usersRes.limit)}
								className="cursor-pointer p-1.5 rounded-lg border border-zinc-800 bg-zinc-900 text-zinc-300 disabled:opacity-40 disabled:cursor-not-allowed hover:bg-zinc-800"
							>
								<ChevronRight className="h-4 w-4" />
							</button>
						</div>
					</div>
				)}
			</div>

			{/* Event Moderation Section */}
			<div className="p-6 rounded-2xl border border-zinc-900 bg-zinc-950/30 backdrop-blur-md space-y-4 text-left">
				<div className="border-b border-zinc-900 pb-3">
					<h2 className="text-xl font-bold text-zinc-200">Event Moderation</h2>
					<p className="text-xs text-zinc-500 mt-1">Review system events. Flag and deactivate spam or reported meetups.</p>
				</div>

				{/* Events List */}
				<div className="space-y-3 max-h-[400px] overflow-y-auto pr-1">
					{events.length > 0 ? (
						events.map((e) => (
							<div key={e.id} className="p-4 rounded-xl border border-zinc-900 bg-zinc-950/50 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
								<div className="space-y-1">
									<div className="flex items-center gap-2.5">
										<h3 className="font-semibold text-zinc-200 text-sm truncate">{e.title}</h3>
										<span className={`text-[9px] font-bold px-2 py-0.5 rounded-full uppercase ${
											e.status === "published"
												? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
												: e.status === "cancelled"
												? "bg-red-500/10 text-red-400 border border-red-500/20"
												: "bg-zinc-800 text-zinc-400"
										}`}>
											{e.status}
										</span>
									</div>
									<p className="text-xs text-zinc-500 flex items-center gap-1.5">
										<MapPin className="h-3 w-3" /> {e.venue_name || "Online"} | Date: {new Date(e.starts_at).toLocaleDateString()}
									</p>
								</div>
								<div className="flex items-center gap-4">
									<div className="text-right text-xs">
										<p className="text-zinc-500">Attendee conversion</p>
										<p className="font-semibold text-zinc-300 mt-0.5">{e.rsvp_count} Confirmed</p>
									</div>
									{e.status !== "cancelled" ? (
										<button
											onClick={() => handleModerateDeleteEvent(e.id, e.title)}
											className="cursor-pointer p-2 rounded-lg border border-red-500/20 bg-red-500/5 text-red-400 hover:bg-red-500/10 transition-colors"
											title="Cancel/moderate event"
										>
											<Trash2 className="h-4.5 w-4.5" />
										</button>
									) : (
										<button
											disabled
											className="p-2 rounded-lg border border-zinc-900 bg-zinc-900/40 text-zinc-600 cursor-not-allowed"
											title="Event already moderated/cancelled"
										>
											<Trash2 className="h-4.5 w-4.5" />
										</button>
									)}
								</div>
							</div>
						))
					) : (
						<p className="text-xs text-zinc-500 text-center py-6">No events currently scheduled on the platform.</p>
					)}
				</div>
			</div>
		</div>
	);
}
