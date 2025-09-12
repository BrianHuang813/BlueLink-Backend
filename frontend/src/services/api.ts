import axios from 'axios';
import { Project, DonationReceipt } from '../types';

const API_BASE_URL = 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

export const projectService = {
  getAllProjects: async (): Promise<Project[]> => {
    const response = await api.get('/projects');
    return response.data;
  },

  getProject: async (id: string): Promise<Project> => {
    const response = await api.get(`/projects/${id}`);
    return response.data;
  },

  getDonationHistory: async (address: string): Promise<DonationReceipt[]> => {
    const response = await api.get(`/donors/${address}`);
    return response.data;
  },
};

export default api;
