"use client";

import { useEffect, useState } from "react";
import { useAuth } from "@clerk/nextjs";
import { api, Event, Community, RSVPWithUser, User } from "../../../lib/api";
import { 
	Calendar, 
	Users, 
	Plus, 
	Send, 
	MapPin, 
	Clock,
	ChevronRight,
	AlertCircle,
	Volume2,
	CheckCircle,
	XCircle
} from "lucide-react";

export default function OrganizerPage() {
	const { getToken, isSignedIn } = useAuth();
	const [currentUser, setCurrentUser] = useState<User | null>(null);
	const [events, setEvents] = useState<Event[]>([]);
	const [communities, setCommunities] = useState<Community[]>([]);
	const [selectedEvent, setSelectedEvent] = useState<Event | null>(null);
	const [rsvps, setRsvps] = useState<RSVPWithUser[]>([]);
	const [interests, setInterests] = useState<{ id: string; name: string }[]>([]);
	const [loading, setLoading] = useState(true);

	// Modals toggles
	const [showEventModal, setShowEventModal] = useState(false);
	const [showCommunityModal, setShowCommunityModal] = useState(false);
	const [showAnnouncementModal, setShowAnnouncementModal] = useState(false);
	const [announcementCommunityId, setAnnouncementCommunityId] = useState("");

	// Forms states
	const [eventForm, setEventForm] = useState({
		title: "",
		description: "",
		venue_name: "",
		venue_address: "",
		starts_at: "",
		ends_at: "",
		max_participants: "",
		age_min: "",
		age_max: "",
		status: "published",
	});

	const [communityForm, setCommunityForm] = useState({
		name: "",
		description: "",
		city: "",
		neighborhood: "",
		max_members: "",
		age_min: "",
		age_max: "",
		visibility: "public",
	});

	const [announcementForm, setAnnouncementForm] = useState({
		title: "",
		body: "",
	});

	const [successMsg, setSuccessMsg] = useState("");
	const [errorMsg, setErrorMsg] = useState("");

	useEffect(() => {
		async function loadData() {
			try {
				const token = await getToken();
				if (token) {
					const [profile, allEvents, allCommunities, interestList] = await Promise.all([
						api.getMe(token),
						api.listEvents(token),
						api.listCommunities(token),
						api.listInterests(token),
					]);
					setCurrentUser(profile);
					setInterests(interestList);

					// Filter events organized by the current user
					const myEvents = allEvents.filter(e => e.organizer_id === profile.id);
					setEvents(myEvents);
					setCommunities(allCommunities);

					if (myEvents.length > 0) {
						setSelectedEvent(myEvents[0]);
					}
				}
			} catch (err) {
				console.error("Failed to load organizer dashboard data:", err);
				setErrorMsg("Failed to retrieve dashboard details.");
			} finally {
				setLoading(false);
			}
		}

		if (isSignedIn) {
			loadData();
		}
	}, [isSignedIn, getToken]);

	// Fetch RSVPs when selectedEvent changes
	useEffect(() => {
		async function fetchRSVPs() {
			if (!selectedEvent) return;
			try {
				const token = await getToken();
				if (token) {
					const list = await api.listEventRSVPs(token, selectedEvent.id);
					setRsvps(list);
				}
			} catch (err) {
				console.error("Failed to fetch RSVPs:", err);
			}
		}
		fetchRSVPs();
	}, [selectedEvent, getToken]);

	const handleCreateEvent = async (e: React.FormEvent) => {
		e.preventDefault();
		setErrorMsg("");
		setSuccessMsg("");

		try {
			const token = await getToken();
			if (!token) return;

			const payload: Partial<Event> = {
				title: eventForm.title,
				description: eventForm.description || null,
				venue_name: eventForm.venue_name || null,
				venue_address: eventForm.venue_address || null,
				starts_at: new Date(eventForm.starts_at).toISOString(),
				ends_at: eventForm.ends_at ? new Date(eventForm.ends_at).toISOString() : null,
				max_participants: eventForm.max_participants ? parseInt(eventForm.max_participants) : null,
				age_min: eventForm.age_min ? parseInt(eventForm.age_min) : null,
				age_max: eventForm.age_max ? parseInt(eventForm.age_max) : null,
				status: eventForm.status,
			};

			const newEvent = await api.createEvent(token, payload);
			setEvents([newEvent, ...events]);
			setSelectedEvent(newEvent);
			setShowEventModal(false);
			setSuccessMsg(`Event "${newEvent.title}" created successfully!`);
			
			// Reset form
			setEventForm({
				title: "",
				description: "",
				venue_name: "",
				venue_address: "",
				starts_at: "",
				ends_at: "",
				max_participants: "",
				age_min: "",
				age_max: "",
				status: "published",
			});
		} catch (err: any) {
			console.error("Failed to create event:", err);
			setErrorMsg(err.message || "Failed to create event. Verify fields.");
		}
	};

	const handleCreateCommunity = async (e: React.FormEvent) => {
		e.preventDefault();
		setErrorMsg("");
		setSuccessMsg("");

		try {
			const token = await getToken();
			if (!token) return;

			const payload: Partial<Community> = {
				name: communityForm.name,
				description: communityForm.description || null,
				city: communityForm.city || null,
				neighborhood: communityForm.neighborhood || null,
				max_members: communityForm.max_members ? parseInt(communityForm.max_members) : null,
				age_min: communityForm.age_min ? parseInt(communityForm.age_min) : null,
				age_max: communityForm.age_max ? parseInt(communityForm.age_max) : null,
				visibility: communityForm.visibility,
			};

			const newCommunity = await api.createCommunity(token, payload);
			setCommunities([newCommunity, ...communities]);
			setShowCommunityModal(false);
			setSuccessMsg(`Community "${newCommunity.name}" created successfully!`);

			// Reset form
			setCommunityForm({
				name: "",
				description: "",
				city: "",
				neighborhood: "",
				max_members: "",
				age_min: "",
				age_max: "",
				visibility: "public",
			});
		} catch (err: any) {
			console.error("Failed to create community:", err);
			setErrorMsg(err.message || "Failed to create community.");
		}
	};

	const handleSendAnnouncement = async (e: React.FormEvent) => {
		e.preventDefault();
		setErrorMsg("");
		setSuccessMsg("");

		try {
			const token = await getToken();
			if (!token) return;

			// Send community announcement request via API
			const response = await fetch(`http://localhost:8080/communities/${announcementCommunityId}/announcements`, {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
					"Authorization": `Bearer ${token}`
				},
				body: JSON.stringify(announcementForm)
			});

			if (!response.ok) {
				throw new Error("Failed to broadcast push notification announcement");
			}

			setShowAnnouncementModal(false);
			setSuccessMsg("Announcement broadcast successfully to all community members!");
			setAnnouncementForm({ title: "", body: "" });
		} catch (err: any) {
			console.error("Announcement error:", err);
			setErrorMsg(err.message || "Failed to send community announcement.");
		}
	};

	if (loading) {
		return (
			<div className="flex flex-col items-center justify-center min-h-[50vh] gap-4">
				<div className="h-8 w-8 border-4 border-violet-600 border-t-transparent rounded-full animate-spin" />
				<p className="text-zinc-500 text-sm">Loading organizer profile and metrics...</p>
			</div>
		);
	}

	return (
		<div className="space-y-8">
			{/* Header Actions */}
			<div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
				<div className="text-left">
					<h1 className="text-3xl font-extrabold tracking-tight">Organizer Dashboard</h1>
					<p className="text-zinc-400 text-sm mt-1">Manage meetups, review attendee counts, and schedule activities.</p>
				</div>
				<div className="flex flex-wrap gap-3">
					<button
						onClick={() => setShowCommunityModal(true)}
						className="cursor-pointer inline-flex items-center gap-2 px-4 py-2 text-sm font-semibold rounded-xl bg-zinc-900 border border-zinc-800 hover:bg-zinc-800 hover:border-zinc-700 transition-colors text-zinc-100"
					>
						<Plus className="h-4 w-4" /> Create Community
					</button>
					<button
						onClick={() => setShowEventModal(true)}
						className="cursor-pointer inline-flex items-center gap-2 px-4 py-2 text-sm font-semibold rounded-xl bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 transition-all text-white shadow-lg shadow-violet-600/15"
					>
						<Plus className="h-4 w-4" /> Create Event
					</button>
				</div>
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
					<AlertCircle className="h-4 w-4 flex-shrink-0" />
					<span>{errorMsg}</span>
				</div>
			)}

			{/* Statistics / RSVP Charts */}
			{events.length > 0 ? (
				<div className="p-6 rounded-2xl border border-zinc-900 bg-zinc-950/30 backdrop-blur-md space-y-6 text-left">
					<div className="flex justify-between items-center">
						<h2 className="text-lg font-bold text-zinc-200">Event RSVP conversion</h2>
						<span className="text-xs text-zinc-500">Attendee response count metrics</span>
					</div>

					{/* Custom SVG Bar Chart */}
					<div className="w-full h-48 flex items-end justify-between gap-4 pt-6 border-b border-zinc-800/60 pb-1">
						{events.slice(0, 5).map((e) => {
							const maxVal = Math.max(...events.map(ev => ev.rsvp_count), 5);
							const percent = (e.rsvp_count / maxVal) * 100;
							return (
								<div key={e.id} className="flex-1 flex flex-col items-center gap-2 h-full justify-end group">
									{/* RSVP Count Label */}
									<span className="text-xs font-semibold text-violet-400 opacity-0 group-hover:opacity-100 transition-opacity bg-zinc-900 border border-zinc-800 px-2 py-0.5 rounded-md mb-1 shadow-md shadow-black/50">
										{e.rsvp_count} RSVP{e.rsvp_count !== 1 ? 's' : ''}
									</span>
									{/* Chart Bar */}
									<div 
										style={{ height: `${percent}%` }}
										className="w-full max-w-[60px] rounded-t-lg bg-gradient-to-t from-violet-600/30 to-violet-500 hover:from-violet-500 hover:to-indigo-500 transition-all duration-500 cursor-pointer shadow-lg shadow-violet-500/5 hover:shadow-violet-500/20 border-t border-violet-400/25"
									/>
									{/* Title Label */}
									<span className="text-[10px] text-zinc-500 truncate max-w-[80px]" title={e.title}>
										{e.title}
									</span>
								</div>
							);
						})}
					</div>
				</div>
			) : (
				<div className="p-12 rounded-2xl border border-zinc-900 bg-zinc-950/10 text-center space-y-3">
					<p className="text-sm text-zinc-500">You haven't organized any events yet.</p>
					<button
						onClick={() => setShowEventModal(true)}
						className="cursor-pointer text-xs font-semibold text-violet-400 hover:text-violet-300 transition-colors"
					>
						Schedule your first event now
					</button>
				</div>
			)}

			{/* Main Grid: Events & RSVP Moderation list */}
			<div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
				{/* Events list */}
				<div className="lg:col-span-1 p-6 rounded-2xl border border-zinc-900 bg-zinc-950/30 backdrop-blur-md space-y-4 text-left">
					<div className="flex justify-between items-center border-b border-zinc-900 pb-3">
						<h2 className="text-lg font-bold text-zinc-200">Scheduled Events</h2>
						<span className="text-xs text-zinc-500">{events.length} Active</span>
					</div>
					
					{events.length > 0 ? (
						<div className="space-y-2 max-h-[450px] overflow-y-auto pr-1">
							{events.map((e) => {
								const active = selectedEvent?.id === e.id;
								return (
									<button
										key={e.id}
										onClick={() => setSelectedEvent(e)}
										className={`cursor-pointer w-full p-4 rounded-xl text-left border transition-all ${
											active 
												? "bg-violet-600/10 border-violet-500/25 shadow-md shadow-violet-500/5"
												: "bg-zinc-900/10 border-zinc-900 hover:bg-zinc-900/30 hover:border-zinc-800"
										}`}
									>
										<p className={`font-semibold text-sm truncate ${active ? "text-violet-300" : "text-zinc-200"}`}>
											{e.title}
										</p>
										<div className="flex items-center gap-1.5 text-xs text-zinc-500 mt-2">
											<Clock className="h-3.5 w-3.5" />
											<span>{new Date(e.starts_at).toLocaleDateString()}</span>
										</div>
										<div className="flex items-center justify-between mt-3 text-xs">
											<span className={`px-2 py-0.5 rounded-full capitalize text-[10px] font-bold ${
												e.status === "published"
													? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
													: "bg-zinc-800 text-zinc-400"
											}`}>
												{e.status}
											</span>
											<span className="text-zinc-400 font-semibold">{e.rsvp_count} RSVP's</span>
										</div>
									</button>
								);
							})}
						</div>
					) : (
						<p className="text-xs text-zinc-500 text-center py-6">No scheduled events.</p>
					)}
				</div>

				{/* RSVPs attendees moderation list */}
				<div className="lg:col-span-2 p-6 rounded-2xl border border-zinc-900 bg-zinc-950/30 backdrop-blur-md space-y-4 text-left flex flex-col justify-between">
					<div className="space-y-4">
						<div className="flex justify-between items-center border-b border-zinc-900 pb-3">
							<div>
								<h2 className="text-lg font-bold text-zinc-200">
									{selectedEvent ? `Attendees: ${selectedEvent.title}` : "Attendee Moderation"}
								</h2>
								{selectedEvent && (
									<p className="text-xs text-zinc-500 mt-1 flex items-center gap-1">
										<MapPin className="h-3 w-3" /> {selectedEvent.venue_name || "No location set"}
									</p>
								)}
							</div>
							<span className="text-xs text-zinc-500">{rsvps.length} RSVP Record(s)</span>
						</div>

						{selectedEvent ? (
							<div className="space-y-3 max-h-[360px] overflow-y-auto pr-1">
								{rsvps.length > 0 ? (
									rsvps.map((r) => (
										<div key={r.id} className="p-4 rounded-xl border border-zinc-900 bg-zinc-950/50 flex items-center justify-between gap-4">
											<div className="flex items-center gap-3 min-w-0">
												<img 
													src={r.user.avatar_url || "https://www.gravatar.com/avatar/00000000000000000000000000000000?d=mp&f=y"}
													alt="User avatar"
													className="h-9 w-9 rounded-full bg-zinc-800 border border-zinc-800"
												/>
												<div className="text-left min-w-0">
													<p className="text-sm font-semibold text-zinc-200 truncate">
														{r.user.full_name || "Flock User"}
													</p>
													<p className="text-xs text-zinc-500 truncate">{r.user.email}</p>
												</div>
											</div>
											<div className="flex items-center gap-3">
												<span className={`text-[10px] font-bold px-2 py-0.5 rounded-full uppercase ${
													r.status === "confirmed"
														? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
														: r.status === "waitlisted"
														? "bg-amber-500/10 text-amber-400 border border-amber-500/20"
														: r.status === "pending"
														? "bg-zinc-800 text-zinc-400"
														: "bg-red-500/10 text-red-400 border border-red-500/20"
												}`}>
													{r.status}
												</span>
											</div>
										</div>
									))
								) : (
									<p className="text-xs text-zinc-500 text-center py-12">No RSVPs registered for this event yet.</p>
								)}
							</div>
						) : (
							<p className="text-xs text-zinc-500 text-center py-12">Select an event from the panel list to manage RSVPs.</p>
						)}
					</div>

					{selectedEvent && (
						<div className="border-t border-zinc-900 pt-4 mt-6 flex justify-between items-center text-xs text-zinc-500">
							<span>Max Attendees limit: {selectedEvent.max_participants || "Unlimited"}</span>
							<span>Age Restrictions: {selectedEvent.age_min || 0} - {selectedEvent.age_max || "Any"} yrs</span>
						</div>
					)}
				</div>
			</div>

			{/* Communities list and Broadcast panel */}
			<div className="p-6 rounded-2xl border border-zinc-900 bg-zinc-950/30 backdrop-blur-md space-y-4 text-left">
				<div className="flex justify-between items-center border-b border-zinc-900 pb-3">
					<h2 className="text-lg font-bold text-zinc-200">Your Communities</h2>
					<span className="text-xs text-zinc-500">{communities.length} Active Group(s)</span>
				</div>
				{communities.length > 0 ? (
					<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
						{communities.map((c) => (
							<div key={c.id} className="p-4 rounded-xl border border-zinc-900 bg-zinc-950/50 flex flex-col justify-between gap-4">
								<div>
									<h3 className="font-bold text-zinc-200 truncate">{c.name}</h3>
									<p className="text-xs text-zinc-500 truncate mt-1">{c.city || "Online"}</p>
									<p className="text-xs text-zinc-400 line-clamp-2 mt-2 leading-relaxed">
										{c.description || "No description set."}
									</p>
								</div>
								<div className="flex justify-between items-center mt-3 pt-3 border-t border-zinc-900 text-xs">
									<span className="text-zinc-500 font-semibold">{c.member_count} member(s)</span>
									<button
										onClick={() => {
											setAnnouncementCommunityId(c.id);
											setShowAnnouncementModal(true);
										}}
										className="cursor-pointer text-xs font-semibold text-violet-400 hover:text-violet-300 transition-colors flex items-center gap-1"
									>
										<Volume2 className="h-3.5 w-3.5" /> Broadcast Alert
									</button>
								</div>
							</div>
						))}
					</div>
				) : (
					<p className="text-xs text-zinc-500 text-center py-6">No community squads found.</p>
				)}
			</div>

			{/* Modal: Create Event */}
			{showEventModal && (
				<div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
					<div className="w-full max-w-lg bg-zinc-950 border border-zinc-900 rounded-3xl p-6 space-y-4 max-h-[90vh] overflow-y-auto text-left shadow-2xl shadow-violet-500/5">
						<h3 className="text-xl font-bold text-white">Create New Event</h3>
						<form onSubmit={handleCreateEvent} className="space-y-4">
							<div className="space-y-1">
								<label className="text-xs text-zinc-400 font-semibold">Event Title *</label>
								<input
									type="text"
									required
									value={eventForm.title}
									onChange={(e) => setEventForm({ ...eventForm, title: e.target.value })}
									placeholder="e.g. Weekly Tech Hackathon"
									className="w-full bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-violet-500"
								/>
							</div>
							<div className="space-y-1">
								<label className="text-xs text-zinc-400 font-semibold">Description</label>
								<textarea
									value={eventForm.description}
									onChange={(e) => setEventForm({ ...eventForm, description: e.target.value })}
									placeholder="Provide event details, schedule, or prerequisites..."
									rows={3}
									className="w-full bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-violet-500 resize-none"
								/>
							</div>
							<div className="grid grid-cols-2 gap-4">
								<div className="space-y-1">
									<label className="text-xs text-zinc-400 font-semibold">Venue Name</label>
									<input
										type="text"
										value={eventForm.venue_name}
										onChange={(e) => setEventForm({ ...eventForm, venue_name: e.target.value })}
										placeholder="e.g. Tech Hub Café"
										className="w-full bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-violet-500"
									/>
								</div>
								<div className="space-y-1">
									<label className="text-xs text-zinc-400 font-semibold">Address</label>
									<input
										type="text"
										value={eventForm.venue_address}
										onChange={(e) => setEventForm({ ...eventForm, venue_address: e.target.value })}
										placeholder="e.g. 123 Main St, Berlin"
										className="w-full bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-violet-500"
									/>
								</div>
							</div>
							<div className="grid grid-cols-2 gap-4">
								<div className="space-y-1">
									<label className="text-xs text-zinc-400 font-semibold">Starts At *</label>
									<input
										type="datetime-local"
										required
										value={eventForm.starts_at}
										onChange={(e) => setEventForm({ ...eventForm, starts_at: e.target.value })}
										className="w-full bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-violet-500"
									/>
								</div>
								<div className="space-y-1">
									<label className="text-xs text-zinc-400 font-semibold">Ends At</label>
									<input
										type="datetime-local"
										value={eventForm.ends_at}
										onChange={(e) => setEventForm({ ...eventForm, ends_at: e.target.value })}
										className="w-full bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-violet-500"
									/>
								</div>
							</div>
							<div className="grid grid-cols-3 gap-4">
								<div className="space-y-1">
									<label className="text-xs text-zinc-400 font-semibold">Max Attendees</label>
									<input
										type="number"
										value={eventForm.max_participants}
										onChange={(e) => setEventForm({ ...eventForm, max_participants: e.target.value })}
										placeholder="e.g. 50"
										className="w-full bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-violet-500"
									/>
								</div>
								<div className="space-y-1">
									<label className="text-xs text-zinc-400 font-semibold">Min Age</label>
									<input
										type="number"
										value={eventForm.age_min}
										onChange={(e) => setEventForm({ ...eventForm, age_min: e.target.value })}
										placeholder="18"
										className="w-full bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-violet-500"
									/>
								</div>
								<div className="space-y-1">
									<label className="text-xs text-zinc-400 font-semibold">Max Age</label>
									<input
										type="number"
										value={eventForm.age_max}
										onChange={(e) => setEventForm({ ...eventForm, age_max: e.target.value })}
										placeholder="30"
										className="w-full bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-violet-500"
									/>
								</div>
							</div>
							<div className="space-y-1">
								<label className="text-xs text-zinc-400 font-semibold">Event Status</label>
								<select
									value={eventForm.status}
									onChange={(e) => setEventForm({ ...eventForm, status: e.target.value })}
									className="w-full bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-violet-500"
								>
									<option value="published">Published (Visible to all users)</option>
									<option value="draft">Draft (Private)</option>
								</select>
							</div>

							<div className="flex justify-end gap-3 pt-4 border-t border-zinc-900">
								<button
									type="button"
									onClick={() => setShowEventModal(false)}
									className="cursor-pointer px-4 py-2 text-sm font-semibold rounded-xl bg-zinc-900 hover:bg-zinc-850 text-zinc-300"
								>
									Cancel
								</button>
								<button
									type="submit"
									className="cursor-pointer px-4 py-2 text-sm font-semibold rounded-xl bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 text-white"
								>
									Create Event
								</button>
							</div>
						</form>
					</div>
				</div>
			)}

			{/* Modal: Create Community */}
			{showCommunityModal && (
				<div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
					<div className="w-full max-w-lg bg-zinc-950 border border-zinc-900 rounded-3xl p-6 space-y-4 max-h-[90vh] overflow-y-auto text-left shadow-2xl shadow-violet-500/5">
						<h3 className="text-xl font-bold text-white">Create New Community</h3>
						<form onSubmit={handleCreateCommunity} className="space-y-4">
							<div className="space-y-1">
								<label className="text-xs text-zinc-400 font-semibold">Community Name *</label>
								<input
									type="text"
									required
									value={communityForm.name}
									onChange={(e) => setCommunityForm({ ...communityForm, name: e.target.value })}
									placeholder="e.g. Berlin Hiking Squad"
									className="w-full bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-violet-500"
								/>
							</div>
							<div className="space-y-1">
								<label className="text-xs text-zinc-400 font-semibold">Description</label>
								<textarea
									value={communityForm.description}
									onChange={(e) => setCommunityForm({ ...communityForm, description: e.target.value })}
									placeholder="What is this community about? What activities are expected? Guidelines..."
									rows={3}
									className="w-full bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-violet-500 resize-none"
								/>
							</div>
							<div className="grid grid-cols-2 gap-4">
								<div className="space-y-1">
									<label className="text-xs text-zinc-400 font-semibold">City</label>
									<input
										type="text"
										value={communityForm.city}
										onChange={(e) => setCommunityForm({ ...communityForm, city: e.target.value })}
										placeholder="e.g. Berlin"
										className="w-full bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-violet-500"
									/>
								</div>
								<div className="space-y-1">
									<label className="text-xs text-zinc-400 font-semibold">Neighborhood</label>
									<input
										type="text"
										value={communityForm.neighborhood}
										onChange={(e) => setCommunityForm({ ...communityForm, neighborhood: e.target.value })}
										placeholder="e.g. Mitte"
										className="w-full bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-violet-500"
									/>
								</div>
							</div>
							<div className="grid grid-cols-3 gap-4">
								<div className="space-y-1">
									<label className="text-xs text-zinc-400 font-semibold">Max Members</label>
									<input
										type="number"
										value={communityForm.max_members}
										onChange={(e) => setCommunityForm({ ...communityForm, max_members: e.target.value })}
										placeholder="e.g. 100"
										className="w-full bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-violet-500"
									/>
								</div>
								<div className="space-y-1">
									<label className="text-xs text-zinc-400 font-semibold">Min Age</label>
									<input
										type="number"
										value={communityForm.age_min}
										onChange={(e) => setCommunityForm({ ...communityForm, age_min: e.target.value })}
										placeholder="18"
										className="w-full bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-violet-500"
									/>
								</div>
								<div className="space-y-1">
									<label className="text-xs text-zinc-400 font-semibold">Max Age</label>
									<input
										type="number"
										value={communityForm.age_max}
										onChange={(e) => setCommunityForm({ ...communityForm, age_max: e.target.value })}
										placeholder="35"
										className="w-full bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-violet-500"
									/>
								</div>
							</div>
							<div className="space-y-1">
								<label className="text-xs text-zinc-400 font-semibold">Visibility</label>
								<select
									value={communityForm.visibility}
									onChange={(e) => setCommunityForm({ ...communityForm, visibility: e.target.value })}
									className="w-full bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-violet-500"
								>
									<option value="public">Public (Visible to everyone)</option>
									<option value="private">Private (Invite only)</option>
								</select>
							</div>

							<div className="flex justify-end gap-3 pt-4 border-t border-zinc-900">
								<button
									type="button"
									onClick={() => setShowCommunityModal(false)}
									className="cursor-pointer px-4 py-2 text-sm font-semibold rounded-xl bg-zinc-900 hover:bg-zinc-850 text-zinc-300"
								>
									Cancel
								</button>
								<button
									type="submit"
									className="cursor-pointer px-4 py-2 text-sm font-semibold rounded-xl bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 text-white"
								>
									Create Community
								</button>
							</div>
						</form>
					</div>
				</div>
			)}

			{/* Modal: Broadcast Announcement */}
			{showAnnouncementModal && (
				<div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
					<div className="w-full max-w-md bg-zinc-950 border border-zinc-900 rounded-3xl p-6 space-y-4 text-left shadow-2xl shadow-violet-500/5">
						<h3 className="text-xl font-bold text-white flex items-center gap-2">
							<Send className="h-5 w-5 text-violet-400" /> Broadcast Community Announcement
						</h3>
						<p className="text-xs text-zinc-500">
							This sends an immediate push notification alert to all registered members of this community.
						</p>
						<form onSubmit={handleSendAnnouncement} className="space-y-4">
							<div className="space-y-1">
								<label className="text-xs text-zinc-400 font-semibold">Announcement Title *</label>
								<input
									type="text"
									required
									value={announcementForm.title}
									onChange={(e) => setAnnouncementForm({ ...announcementForm, title: e.target.value })}
									placeholder="e.g. Schedule Update or Urgent Alert"
									className="w-full bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-violet-500"
								/>
							</div>
							<div className="space-y-1">
								<label className="text-xs text-zinc-400 font-semibold">Body Content *</label>
								<textarea
									required
									value={announcementForm.body}
									onChange={(e) => setAnnouncementForm({ ...announcementForm, body: e.target.value })}
									placeholder="Write your broadcast message alert details here..."
									rows={4}
									className="w-full bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-violet-500 resize-none"
								/>
							</div>

							<div className="flex justify-end gap-3 pt-4 border-t border-zinc-900">
								<button
									type="button"
									onClick={() => setShowAnnouncementModal(false)}
									className="cursor-pointer px-4 py-2 text-sm font-semibold rounded-xl bg-zinc-900 hover:bg-zinc-850 text-zinc-300"
								>
									Cancel
								</button>
								<button
									type="submit"
									className="cursor-pointer px-4 py-2 text-sm font-semibold rounded-xl bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 text-white flex items-center gap-1.5"
								>
									<Send className="h-3.5 w-3.5" /> Broadcast Push
								</button>
							</div>
						</form>
					</div>
				</div>
			)}
		</div>
	);
}
