import React, { useState, useEffect } from 'react';
import { useCurrentAccount } from '@mysten/dapp-kit';
import { projectService } from '../services/api';
import { DonationReceipt } from '../types';

const DashboardPage: React.FC = () => {
  const [donations, setDonations] = useState<DonationReceipt[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const currentAccount = useCurrentAccount();

  useEffect(() => {
    const fetchDonations = async () => {
      if (!currentAccount) {
        setLoading(false);
        return;
      }

      try {
        const donationsData = await projectService.getDonationHistory(currentAccount.address);
        setDonations(donationsData);
      } catch (err) {
        setError('無法加載捐贈記錄');
        console.error('Error fetching donations:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchDonations();
  }, [currentAccount]);

  if (!currentAccount) {
    return (
      <div className="max-w-4xl mx-auto">
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-8 text-center">
          <div className="text-6xl mb-4">👤</div>
          <h2 className="text-2xl font-bold text-yellow-800 mb-2">
            需要連接錢包
          </h2>
          <p className="text-yellow-700 mb-4">
            請連接您的 Sui 錢包以查看您的捐贈記錄
          </p>
          <p className="text-sm text-yellow-600">
            連接錢包後，您可以查看所有捐贈記錄和 NFT 憑證
          </p>
        </div>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-64">
        <div className="text-lg">正在加載捐贈記錄...</div>
      </div>
    );
  }

  const totalDonations = donations.length;
  const totalAmount = donations.reduce((sum, donation) => {
    return sum + (parseFloat(donation.amount) / 1000000000); // Convert MIST to SUI
  }, 0);

  return (
    <div className="max-w-6xl mx-auto">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-800 mb-2">我的儀表板</h1>
        <p className="text-gray-600">查看您的捐贈記錄和 NFT 憑證</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div className="bg-white rounded-lg shadow-md p-6">
          <div className="flex items-center">
            <div className="text-3xl mr-4">🎯</div>
            <div>
              <div className="text-2xl font-bold text-blue-600">{totalDonations}</div>
              <div className="text-sm text-gray-600">支持的項目</div>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-md p-6">
          <div className="flex items-center">
            <div className="text-3xl mr-4">💝</div>
            <div>
              <div className="text-2xl font-bold text-green-600">
                {totalAmount.toFixed(2)} SUI
              </div>
              <div className="text-sm text-gray-600">總捐贈金額</div>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-md p-6">
          <div className="flex items-center">
            <div className="text-3xl mr-4">🏆</div>
            <div>
              <div className="text-2xl font-bold text-purple-600">{totalDonations}</div>
              <div className="text-sm text-gray-600">NFT 憑證</div>
            </div>
          </div>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow-lg overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-xl font-semibold text-gray-800">捐贈記錄</h2>
        </div>

        {error ? (
          <div className="p-6">
            <div className="text-red-600 text-center">{error}</div>
          </div>
        ) : donations.length === 0 ? (
          <div className="p-12 text-center">
            <div className="text-6xl mb-4">🌱</div>
            <h3 className="text-xl font-semibold text-gray-700 mb-2">
              還沒有捐贈記錄
            </h3>
            <p className="text-gray-600 mb-6">
              開始支持可持續發展項目，獲得您的第一個 NFT 憑證！
            </p>
            <a 
              href="/" 
              className="bg-blue-600 text-white px-6 py-3 rounded-lg hover:bg-blue-700 transition-colors inline-block"
            >
              探索項目
            </a>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    NFT 憑證 ID
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    項目 ID
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    捐贈金額
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    操作
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {donations.map((donation) => (
                  <tr key={donation.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm font-medium text-gray-900">
                        {donation.id.substring(0, 12)}...
                      </div>
                      <div className="text-xs text-gray-500 font-mono">
                        NFT 憑證
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm text-gray-900 font-mono">
                        {donation.project_id.substring(0, 12)}...
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm font-bold text-green-600">
                        {(parseFloat(donation.amount) / 1000000000).toFixed(4)} SUI
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                      <a 
                        href={`/project/${donation.project_id}`}
                        className="text-blue-600 hover:text-blue-800 hover:underline"
                      >
                        查看項目
                      </a>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <div className="mt-8 bg-blue-50 border border-blue-200 rounded-lg p-6">
        <h3 className="text-lg font-semibold text-blue-800 mb-3">
          📖 關於 NFT 捐贈憑證
        </h3>
        <div className="text-blue-700 space-y-2">
          <p>• 每次捐贈都會為您鑄造一個獨特的 NFT 憑證</p>
          <p>• NFT 憑證永久保存在區塊鏈上，無法偽造</p>
          <p>• 您可以在 Sui 錢包中查看和管理這些 NFT</p>
          <p>• 憑證包含捐贈金額、項目信息和時間戳記</p>
        </div>
      </div>

      <div className="mt-6 pt-6 border-t border-gray-200">
        <div className="text-sm text-gray-600">
          <strong>當前錢包地址：</strong>
          <div className="font-mono text-xs mt-1 break-all">
            {currentAccount.address}
          </div>
        </div>
      </div>
    </div>
  );
};

export default DashboardPage;
