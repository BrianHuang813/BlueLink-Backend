export interface Project {
  id: string;
  creator: string;
  name: string;
  description: string;
  funding_goal: string;
  total_raised: string;
  donor_count: string;
}

export interface DonationReceipt {
  id: string;
  project_id: string;
  donor: string;
  amount: string;
}

export interface CreateProjectForm {
  name: string;
  description: string;
  funding_goal: number;
}

export interface DonationForm {
  amount: number;
}
