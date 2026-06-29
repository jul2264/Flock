const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface User {
	id: string;
	clerk_id: string;
	email: string;
	full_name: string | null;
	username: string | null;
	avatar_url: string | null;
	date_of_birth: string | null;
	city: string | null;
	neighborhood: string | null;
	latitude: number | null;
	longitude: number | null;
	search_radius: number;
	role: string;
	last_seen_at: string | null;
	created_at: string;
	updated_at: string;
}

export interface Event {
	id: string;
	organizer_id: string;
	community_id: string | null;
	title: string;
	description: string | null;
	venue_name: string | null;
	venue_address: string | null;
	google_place_id: string | null;
	latitude: number | null;
	longitude: number | null;
	starts_at: string;
	ends_at: string | null;
	max_participants: number | null;
	rsvp_count: number;
	age_min: number | null;
	age_max: number | null;
	status: string;
	banner_url: string | null;
	created_at: string;
	updated_at: string;
}

export interface Community {
	id: string;
	name: string;
	description: string | null;
	city: string | null;
	neighborhood: string | null;
	latitude: number | null;
	longitude: number | null;
	is_recurring: boolean;
	recurrence_rule: string | null;
	max_members: number | null;
	member_count: number;
	age_min: number | null;
	age_max: number | null;
	visibility: string;
	status: string;
	image_url: string | null;
	created_at: string;
	updated_at: string;
}

export interface RSVP {
	id: string;
	event_id: string;
	user_id: string;
	status: string;
	attended: boolean;
	created_at: string;
	updated_at: string;
}

export interface RSVPWithUser {
	id: string;
	event_id: string;
	user_id: string;
	status: string;
	attended: boolean;
	created_at: string;
	user: {
		id: string;
		full_name: string | null;
		email: string;
		avatar_url: string | null;
		role: string;
	};
}

export interface PlatformStats {
	total_users: number;
	users_by_role: Record<string, number>;
	active_users_24h: number;
	total_events: number;
	events_by_status: Record<string, number>;
	total_communities: number;
	communities_by_status: Record<string, number>;
	total_rsvps: number;
	rsvps_by_status: Record<string, number>;
}

export interface UsersListResponse {
	users: User[];
	total: number;
	page: number;
	limit: number;
}

async function request<T>(path: string, token?: string, options: RequestInit = {}): Promise<T> {
	const headers = new Headers(options.headers || {});
	headers.set("Content-Type", "application/json");
	if (token) {
		headers.set("Authorization", `Bearer ${token}`);
	}

	const response = await fetch(`${API_BASE_URL}${path}`, {
		...options,
		headers,
	});

	if (!response.ok) {
		const text = await response.text();
		throw new Error(text || `Request failed with status ${response.status}`);
	}

	if (response.status === 204) {
		return {} as T;
	}

	return response.json();
}

export const api = {
	// Sync user profile upon Clerk login
	syncUser: (token: string, email: string, fullName: string, avatarUrl: string): Promise<User> => {
		return request<User>("/users/sync", token, {
			method: "POST",
			body: JSON.stringify({ email, full_name: fullName, avatar_url: avatarUrl }),
		});
	},

	// Get current user details
	getMe: (token: string): Promise<User> => {
		return request<User>("/users/me", token);
	},

	// Events API
	listEvents: (token: string, params: Record<string, string | number> = {}): Promise<Event[]> => {
		const query = new URLSearchParams();
		Object.entries(params).forEach(([key, val]) => {
			if (val !== undefined && val !== null) {
				query.append(key, String(val));
			}
		});
		const queryString = query.toString() ? `?${query.toString()}` : "";
		return request<Event[]>(`/events${queryString}`, token);
	},

	getEvent: (token: string, id: string): Promise<Event> => {
		return request<Event>(`/events/${id}`, token);
	},

	createEvent: (token: string, data: Partial<Event>): Promise<Event> => {
		return request<Event>("/events", token, {
			method: "POST",
			body: JSON.stringify(data),
		});
	},

	updateEvent: (token: string, id: string, data: Partial<Event>): Promise<Event> => {
		return request<Event>(`/events/${id}`, token, {
			method: "PATCH",
			body: JSON.stringify(data),
		});
	},

	deleteEvent: (token: string, id: string): Promise<void> => {
		return request<void>(`/events/${id}`, token, {
			method: "DELETE",
		});
	},

	// Event RSVPs
	listEventRSVPs: (token: string, eventId: string): Promise<RSVPWithUser[]> => {
		return request<RSVPWithUser[]>(`/events/${eventId}/rsvps`, token);
	},

	updateRSVP: (token: string, eventId: string, status: string): Promise<RSVP> => {
		return request<RSVP>(`/events/${eventId}/rsvp`, token, {
			method: "POST",
			body: JSON.stringify({ status }),
		});
	},

	// Communities API
	listCommunities: (token: string): Promise<Community[]> => {
		return request<Community[]>("/communities", token);
	},

	getCommunity: (token: string, id: string): Promise<Community> => {
		return request<Community>(`/communities/${id}`, token);
	},

	createCommunity: (token: string, data: Partial<Community>): Promise<Community> => {
		return request<Community>("/communities", token, {
			method: "POST",
			body: JSON.stringify(data),
		});
	},

	updateCommunity: (token: string, id: string, data: Partial<Community>): Promise<Community> => {
		return request<Community>(`/communities/${id}`, token, {
			method: "PATCH",
			body: JSON.stringify(data),
		});
	},

	deleteCommunity: (token: string, id: string): Promise<void> => {
		return request<void>(`/communities/${id}`, token, {
			method: "DELETE",
		});
	},

	// Interests API
	listInterests: (token: string): Promise<{ id: string; name: string }[]> => {
		return request<{ id: string; name: string }[]>("/interests", token);
	},

	// Admin API
	listAdminUsers: (token: string, page = 1, limit = 20): Promise<UsersListResponse> => {
		return request<UsersListResponse>(`/admin/users?page=${page}&limit=${limit}`, token);
	},

	updateUserRole: (token: string, userId: string, role: string): Promise<User> => {
		return request<User>(`/admin/users/${userId}/role`, token, {
			method: "PATCH",
			body: JSON.stringify({ role }),
		});
	},

	moderateDeleteEvent: (token: string, eventId: string): Promise<void> => {
		return request<void>(`/admin/events/${eventId}`, token, {
			method: "DELETE",
		});
	},

	getAdminStats: (token: string): Promise<PlatformStats> => {
		return request<PlatformStats>("/admin/stats", token);
	},
};
