import React from 'react';
import { useParams } from 'react-router-dom';

const FundDetails: React.FC = () => {
  const { id } = useParams<{ id: string }>();

  return (
    <div className="p-6">
      <h1 className="text-3xl font-bold text-text-primary mb-6">
        Детали фонда #{id}
      </h1>
      <div className="card">
        <p className="text-text-secondary">Страница в разработке...</p>
      </div>
    </div>
  );
};

export default FundDetails;
