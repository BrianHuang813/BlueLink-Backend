import React from 'react';
import { Link } from 'react-router-dom';
import { ConnectButton } from '@mysten/dapp-kit';

const Header: React.FC = () => {
  return (
    <header className="bg-blue-600 text-white shadow-lg">
      <div className="container mx-auto px-4 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-8">
            <Link to="/" className="text-2xl font-bold">
              🔗 BlueLink
            </Link>
            <nav className="hidden md:flex space-x-6">
              <Link 
                to="/" 
                className="hover:text-blue-200 transition-colors"
              >
                首頁
              </Link>
              <Link 
                to="/create" 
                className="hover:text-blue-200 transition-colors"
              >
                創建項目
              </Link>
              <Link 
                to="/dashboard" 
                className="hover:text-blue-200 transition-colors"
              >
                我的儀表板
              </Link>
            </nav>
          </div>
          <ConnectButton />
        </div>
      </div>
    </header>
  );
};

export default Header;
