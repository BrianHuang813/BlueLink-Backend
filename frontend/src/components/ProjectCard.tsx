import React from 'react';
import { Link } from 'react-router-dom';
import { Project } from '../types';

interface ProjectCardProps {
  project: Project;
}

const ProjectCard: React.FC<ProjectCardProps> = ({ project }) => {
  const fundingGoal = parseFloat(project.funding_goal) / 1000000000; // Convert MIST to SUI
  const totalRaised = parseFloat(project.total_raised) / 1000000000; // Convert MIST to SUI
  const progressPercentage = fundingGoal > 0 ? (totalRaised / fundingGoal) * 100 : 0;

  return (
    <div className="bg-white rounded-lg shadow-md p-6 hover:shadow-lg transition-shadow">
      <h3 className="text-xl font-bold mb-2 text-gray-800">{project.name}</h3>
      <p className="text-gray-600 mb-4 line-clamp-3">{project.description}</p>
      
      <div className="mb-4">
        <div className="flex justify-between text-sm text-gray-600 mb-1">
          <span>進度</span>
          <span>{progressPercentage.toFixed(1)}%</span>
        </div>
        <div className="w-full bg-gray-200 rounded-full h-2">
          <div 
            className="bg-blue-600 h-2 rounded-full transition-all duration-300" 
            style={{ width: `${Math.min(progressPercentage, 100)}%` }}
          ></div>
        </div>
      </div>

      <div className="flex justify-between items-center mb-4">
        <div>
          <p className="text-sm text-gray-600">已籌集</p>
          <p className="font-bold text-lg">{totalRaised.toFixed(2)} SUI</p>
        </div>
        <div>
          <p className="text-sm text-gray-600">目標</p>
          <p className="font-bold text-lg">{fundingGoal.toFixed(2)} SUI</p>
        </div>
        <div>
          <p className="text-sm text-gray-600">捐贈者</p>
          <p className="font-bold text-lg">{project.donor_count}</p>
        </div>
      </div>

      <Link 
        to={`/project/${project.id}`}
        className="w-full bg-blue-600 text-white py-2 px-4 rounded-lg hover:bg-blue-700 transition-colors block text-center"
      >
        查看詳情
      </Link>
    </div>
  );
};

export default ProjectCard;
