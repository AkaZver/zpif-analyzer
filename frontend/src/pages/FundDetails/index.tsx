import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Typography, Card, Row, Col, Statistic, Tag, Button, Space, Table,
  message, Spin, Upload, Descriptions, List, Modal, Form, Input, Select,
  Switch, Popconfirm, DatePicker,
} from 'antd';
import { ArrowLeftOutlined, SearchOutlined, UploadOutlined, ThunderboltOutlined, CheckCircleOutlined, CloseCircleOutlined, DeleteOutlined, DownloadOutlined, EditOutlined, CloudDownloadOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import {
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer,
  BarChart, Bar, ReferenceLine,
} from 'recharts';
import { apiClient } from '../../api/client';
import type { Fund, FundFinancials, FundDocument, LLMAnalysis } from '../../types';

const FundDetails: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [fund, setFund] = useState<Fund | null>(null);
  const [financials, setFinancials] = useState<FundFinancials[]>([]);
  const [documents, setDocuments] = useState<FundDocument[]>([]);
  const [analysis, setAnalysis] = useState<LLMAnalysis | null>(null);
  const [loading, setLoading] = useState(true);
  const [analyzing, setAnalyzing] = useState(false);
  const [discovering, setDiscovering] = useState(false);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [editForm] = Form.useForm();
  const [saving, setSaving] = useState(false);
  const [fetchingMarketData, setFetchingMarketData] = useState(false);

  useEffect(() => {
    if (id) loadData();
  }, [id]);

  const loadData = async () => {
    if (!id) return;
    setLoading(true);
    try {
      const fundId = parseInt(id);
      const [fundData, financialsData, documentsData] = await Promise.all([
        apiClient.getFund(fundId),
        apiClient.getFinancials(fundId),
        apiClient.getDocuments(fundId),
      ]);
      setFund(fundData);
      setFinancials(financialsData);
      setDocuments(documentsData);
      try {
        const analysisData = await apiClient.getAnalysis(fundId);
        setAnalysis(analysisData);
      } catch {
        setAnalysis(null);
      }
    } catch {
      message.error('Не удалось загрузить данные фонда');
    } finally {
      setLoading(false);
    }
  };

  const handleDiscover = async () => {
    if (!id) return;
    setDiscovering(true);
    try {
      await apiClient.discoverDocuments(parseInt(id));
      message.success('Автопоиск документов запущен');
      setTimeout(() => loadData(), 3000);
    } catch {
      message.error('Ошибка при поиске документов');
    } finally {
      setDiscovering(false);
    }
  };

  const handleUpload = async (file: File) => {
    if (!id) return false;
    try {
      await apiClient.uploadDocument(parseInt(id), file);
      message.success('Документ загружен');
      await loadData();
    } catch {
      message.error('Ошибка при загрузке');
    }
    return false;
  };

  const handleAnalyze = async () => {
    if (!id) return;
    setAnalyzing(true);
    try {
      const result = await apiClient.analyzeFund(parseInt(id));
      setAnalysis(result);
      message.success('Анализ завершён');
    } catch {
      message.error('Ошибка при анализе');
    } finally {
      setAnalyzing(false);
    }
  };

  const handleDeleteDocument = async (docId: number) => {
    if (!id) return;
    try {
      await apiClient.deleteDocument(parseInt(id), docId);
      message.success('Документ удалён');
      await loadData();
    } catch {
      message.error('Ошибка при удалении');
    }
  };

  const handleDownloadDocument = async (docId: number, fileName: string) => {
    if (!id) return;
    try {
      const blob = await apiClient.downloadDocument(parseInt(id), docId);
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = fileName;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch {
      message.error('Ошибка при скачивании');
    }
  };

  const handleEditFund = () => {
    if (!fund) return;
    editForm.setFieldsValue({
      ...fund,
      fund_start_date: fund.fund_start_date ? dayjs(fund.fund_start_date) : null,
      fund_end_date: fund.fund_end_date ? dayjs(fund.fund_end_date) : null,
    });
    setEditModalVisible(true);
  };

  const handleSaveFund = async () => {
    if (!id) return;
    setSaving(true);
    try {
      const values = await editForm.validateFields();
      const data = {
        ...values,
        fund_start_date: values.fund_start_date ? values.fund_start_date.toISOString() : null,
        fund_end_date: values.fund_end_date ? values.fund_end_date.toISOString() : null,
      };
      await apiClient.updateFund(parseInt(id), data);
      message.success('Фонд обновлён');
      setEditModalVisible(false);
      await loadData();
    } catch {
      message.error('Ошибка при сохранении');
    } finally {
      setSaving(false);
    }
  };

  const handleDeleteFund = async () => {
    if (!id) return;
    try {
      await apiClient.deleteFund(parseInt(id));
      message.success('Фонд удалён');
      navigate('/');
    } catch {
      message.error('Ошибка при удалении');
    }
  };

  const handleFetchMarketData = async () => {
    if (!id) return;
    setFetchingMarketData(true);
    try {
      const result = await apiClient.fetchMarketData(parseInt(id));
      const msg = `Создано: ${result.records_created}, Обновлено: ${result.records_updated}`;
      if (result.moex_available && result.investfunds_available) {
        message.success(`Данные обновлены. ${msg}`);
      } else if (result.moex_available || result.investfunds_available) {
        message.warning(`Частичное обновление. ${msg}`);
      } else {
        message.error('Не удалось получить данные из источников');
      }
      await loadData();
    } catch (error: any) {
      message.error(error?.response?.data?.error || 'Ошибка при загрузке данных');
    } finally {
      setFetchingMarketData(false);
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <Spin size="large" />
      </div>
    );
  }

  if (!fund) {
    return <div className="text-text-primary">Фонд не найден</div>;
  }

  const latest = financials.length > 0 ? (() => {
    // Найти последнюю запись с ненулевым NAV
    const withNav = financials.find(f => f.nav_per_unit_rub > 0);
    // Найти последнюю запись с ненулевым СЧА
    const withSCA = financials.find(f => f.nav_total_mln_rub > 0);
    
    return {
      ...financials[0],
      nav_per_unit_rub: withNav?.nav_per_unit_rub || financials[0].nav_per_unit_rub,
      nav_total_mln_rub: withSCA?.nav_total_mln_rub || financials[0].nav_total_mln_rub,
    };
  })() : null;

  // Найти первую дату с ненулевой ценой пая (начало торгов)
  const firstTradingDate = financials
    .filter(f => f.unit_price_rub > 0)
    .sort((a, b) => new Date(a.snapshot_date).getTime() - new Date(b.snapshot_date).getTime())[0];

  const tradingStartDate = firstTradingDate 
    ? new Date(firstTradingDate.snapshot_date)
    : null;

  const tradingStartFormatted = tradingStartDate
    ? tradingStartDate.toLocaleDateString('ru-RU', { month: 'short', year: '2-digit' })
    : null;

  const priceChartData = (() => {
    // Группируем по месяцам, берём последнюю запись каждого месяца
    const grouped = new Map<string, FundFinancials>();
    
    financials.forEach((f) => {
      // Пропускаем записи с нулевыми значениями
      if (f.unit_price_rub === 0 && f.nav_per_unit_rub === 0) {
        return;
      }
      
      const date = new Date(f.snapshot_date);
      const key = `${date.getFullYear()}-${date.getMonth()}`;
      
      // Берём последнюю запись каждого месяца
      const existing = grouped.get(key);
      if (!existing || new Date(f.snapshot_date) > new Date(existing.snapshot_date)) {
        grouped.set(key, f);
      }
    });
    
    // Сортируем по дате и форматируем
    return Array.from(grouped.values())
      .sort((a, b) => new Date(a.snapshot_date).getTime() - new Date(b.snapshot_date).getTime())
      .map((f) => {
        const currentDate = new Date(f.snapshot_date);
        const formattedDate = currentDate.toLocaleDateString('ru-RU', { month: 'short', year: '2-digit' });
        
        // Показываем цену пая только после начала торгов
        const showPrice = tradingStartDate && currentDate >= tradingStartDate;
        
        return {
          date: formattedDate,
          'Цена пая': showPrice ? f.unit_price_rub : null,
          'РСП': f.nav_per_unit_rub,
        };
      });
  })();

  const payoutChartData = financials
    .filter((f) => f.annual_payout_rub > 0)
    .sort((a, b) => new Date(a.snapshot_date).getTime() - new Date(b.snapshot_date).getTime())
    .map((f) => ({
      date: new Date(f.snapshot_date).toLocaleDateString('ru-RU', { month: 'short', year: '2-digit' }),
      'Выплата': f.annual_payout_rub,
    }));

  const docColumns = [
    { title: 'Файл', dataIndex: 'file_name', key: 'file_name' },
    { title: 'Тип', dataIndex: 'document_type', key: 'document_type' },
    {
      title: 'Размер',
      dataIndex: 'file_size',
      key: 'file_size',
      render: (size: number) => size > 0 ? `${(size / 1024).toFixed(1)} КБ` : '—',
    },
    {
      title: 'Источник',
      dataIndex: 'source',
      key: 'source',
      render: (v: string) => <Tag color={v === 'auto' ? 'blue' : 'default'}>{v === 'auto' ? 'Авто' : 'Ручная'}</Tag>,
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (v: string) => {
        const colors: Record<string, string> = { pending: 'gold', downloaded: 'blue', analyzed: 'green', error: 'red' };
        const labels: Record<string, string> = { pending: 'Ожидает', downloaded: 'Скачан', analyzed: 'Проанализирован', error: 'Ошибка' };
        return <Tag color={colors[v] || 'default'}>{labels[v] || v}</Tag>;
      },
    },
    {
      title: 'Дата',
      dataIndex: 'upload_date',
      key: 'upload_date',
      render: (v: string) => new Date(v).toLocaleDateString('ru-RU'),
    },
    {
      title: 'Действия',
      key: 'actions',
      width: 120,
      render: (_: unknown, record: FundDocument) => (
        <Space>
          <Button type="text" size="small" icon={<DownloadOutlined />} onClick={() => handleDownloadDocument(record.id, record.file_name)} />
          <Button type="text" danger size="small" icon={<DeleteOutlined />} onClick={() => handleDeleteDocument(record.id)} />
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Button
        type="text"
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate('/')}
        className="text-text-primary mb-4"
      >
        Назад к сравнению
      </Button>

      <div className="flex items-center gap-4 mb-6">
        <Typography.Title level={3} className="text-text-primary m-0">
          {fund.name}
        </Typography.Title>
        <Space>
          <Tag>
            {fund.isin}
            {fund.ticker && ` (${fund.ticker})`}
          </Tag>
          {fund.qualified_required && <Tag color="red">Только для квалов</Tag>}
          {fund.has_market_maker && <Tag color="green">Маркет-мейкер</Tag>}
        </Space>
        <Space className="ml-auto">
          <Button icon={<CloudDownloadOutlined />} onClick={handleFetchMarketData} loading={fetchingMarketData}>
            Обновить данные
          </Button>
          <Button icon={<EditOutlined />} onClick={handleEditFund}>
            Редактировать
          </Button>
          <Popconfirm title="Удалить фонд?" onConfirm={handleDeleteFund} okText="Да" cancelText="Нет">
            <Button danger icon={<DeleteOutlined />}>
              Удалить
            </Button>
          </Popconfirm>
        </Space>
      </div>

      <Descriptions className="mb-6" bordered size="small" column={3}>
        <Descriptions.Item label="УК">{fund.management_company || '—'}</Descriptions.Item>
        <Descriptions.Item label="Сегмент">{fund.real_estate_segment || '—'}</Descriptions.Item>
        <Descriptions.Item label="Дата старта">
          {fund.fund_start_date ? new Date(fund.fund_start_date).toLocaleDateString('ru-RU') : '—'}
        </Descriptions.Item>
        <Descriptions.Item label="Дата завершения">
          {fund.fund_end_date ? new Date(fund.fund_end_date).toLocaleDateString('ru-RU') : '—'}
        </Descriptions.Item>
      </Descriptions>

      <Typography.Title level={4} className="text-text-primary mb-4">
        Ключевые метрики
      </Typography.Title>
      <Row gutter={[16, 16]} className="mb-6">
        <Col xs={12} sm={8} md={6}>
          <Card className="bg-surface-card border-0">
            <Statistic title="Цена пая" value={latest?.unit_price_rub || 0} suffix="₽" precision={0} />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card className="bg-surface-card border-0">
            <Statistic title="РСП" value={latest?.nav_per_unit_rub || 0} suffix="₽" precision={0} />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card className="bg-surface-card border-0">
            <Statistic
              title="Дисконт к РСП"
              value={latest?.discount_to_nav_pct || 0}
              suffix="%"
              precision={1}
              valueStyle={{ color: (latest?.discount_to_nav_pct || 0) <= 0 ? '#52c41a' : '#ff4d4f' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card className="bg-surface-card border-0">
            <Statistic title="Cap Rate" value={latest?.cap_rate_pct || 0} suffix="%" precision={1} />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card className="bg-surface-card border-0">
            <Statistic 
              title="СЧА" 
              value={latest?.nav_total_mln_rub || 0} 
              suffix="млн ₽" 
              precision={2} 
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card className="bg-surface-card border-0">
            <Statistic title="P/NAV" value={latest?.p_nav || 0} precision={2} />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card className="bg-surface-card border-0">
            <Statistic title="P/AFFO" value={latest?.p_affo || 0} precision={2} />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card className="bg-surface-card border-0">
            <Statistic
              title="Доходность выплат"
              value={latest?.payout_yield_pct || 0}
              suffix="%"
              precision={1}
              valueStyle={{ color: (latest?.payout_yield_pct || 0) >= 0 ? '#52c41a' : '#ff4d4f' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card className="bg-surface-card border-0">
            <Statistic
              title="Полная доходность"
              value={latest?.total_return_pct || 0}
              suffix="%"
              precision={1}
              valueStyle={{ color: (latest?.total_return_pct || 0) >= 0 ? '#52c41a' : '#ff4d4f' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card className="bg-surface-card border-0">
            <Statistic title="Долг/СЧА" value={latest?.debt_to_nav_ratio || 0} precision={2} />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card className="bg-surface-card border-0">
            <Statistic title="Комиссия УК" value={latest?.management_fee_pct || 0} suffix="%" precision={1} />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card className="bg-surface-card border-0">
            <Statistic title="Объектов" value={latest?.number_of_properties || 0} />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card className="bg-surface-card border-0">
            <Statistic title="Прогноз IRR" value={latest?.irr_forecast_pct || 0} suffix="%" precision={1} />
          </Card>
        </Col>
      </Row>

      {priceChartData.length > 0 && (
        <>
          <Typography.Title level={4} className="text-text-primary mb-4">
            Динамика цены и РСП
          </Typography.Title>
          <Card className="bg-surface-card border-0 mb-6">
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={priceChartData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#444444" />
                <XAxis dataKey="date" stroke="#a0a0a0" />
                <YAxis stroke="#a0a0a0" />
                <Tooltip contentStyle={{ backgroundColor: '#333333', border: 'none' }} />
                <Legend />
                
                {/* Вертикальная линия "Начало торгов" */}
                {tradingStartFormatted && (
                  <ReferenceLine 
                    x={tradingStartFormatted} 
                    stroke="#888888" 
                    strokeDasharray="3 3"
                    label={{ 
                      value: 'Начало торгов', 
                      position: 'top',
                      fill: '#a0a0a0',
                      fontSize: 12,
                    }}
                  />
                )}
                
                <Line 
                  type="monotone" 
                  dataKey="Цена пая" 
                  stroke="#7c5cbf" 
                  strokeWidth={2}
                  connectNulls={false}
                />
                <Line 
                  type="monotone" 
                  dataKey="РСП" 
                  stroke="#e94560" 
                  strokeWidth={2}
                />
              </LineChart>
            </ResponsiveContainer>
          </Card>
        </>
      )}

      {payoutChartData.length > 0 && (
        <>
          <Typography.Title level={4} className="text-text-primary mb-4">
            История выплат
          </Typography.Title>
          <Card className="bg-surface-card border-0 mb-6">
            <ResponsiveContainer width="100%" height={250}>
              <BarChart data={payoutChartData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#444444" />
                <XAxis dataKey="date" stroke="#a0a0a0" />
                <YAxis stroke="#a0a0a0" />
                <Tooltip 
                  contentStyle={{ backgroundColor: '#333333', border: 'none' }}
                  cursor={{ fill: '#444444', fillOpacity: 0.3 }}
                />
                <Bar dataKey="Выплата" fill="#7c5cbf" />
              </BarChart>
            </ResponsiveContainer>
          </Card>
        </>
      )}

      <Typography.Title level={4} className="text-text-primary mb-4">
        Документы
      </Typography.Title>
      <Card className="bg-surface-card border-0 mb-6">
        <Space className="mb-4">
          <Button icon={<SearchOutlined />} onClick={handleDiscover} loading={discovering}>
            Найти в интернете
          </Button>
          <Upload beforeUpload={handleUpload} showUploadList={false} accept=".pdf,.doc,.docx,.xlsx">
            <Button icon={<UploadOutlined />}>Загрузить вручную</Button>
          </Upload>
          <Button
            type="primary"
            icon={<ThunderboltOutlined />}
            onClick={handleAnalyze}
            loading={analyzing}
            disabled={documents.length === 0}
          >
            Запустить анализ
          </Button>
        </Space>
        <Table
          columns={docColumns}
          dataSource={documents}
          rowKey="id"
          pagination={false}
          size="small"
        />
      </Card>

      {analysis && (
        <>
          <Typography.Title level={4} className="text-text-primary mb-4">
            LLM-анализ
          </Typography.Title>
          <Card className="bg-surface-card border-0">
            <Descriptions bordered size="small" column={1} className="mb-4">
              <Descriptions.Item label="Модель">{analysis.model_used}</Descriptions.Item>
              <Descriptions.Item label="Дата">
                {new Date(analysis.created_at).toLocaleString('ru-RU')}
              </Descriptions.Item>
              <Descriptions.Item label="Резюме">{analysis.analysis_summary || '—'}</Descriptions.Item>
              <Descriptions.Item label="Оценка рисков">{analysis.risk_assessment || '—'}</Descriptions.Item>
            </Descriptions>
            
            {(() => {
              try {
                const prosCons = JSON.parse(analysis.pros_cons || '{}');
                const pros = prosCons.pros || [];
                const cons = prosCons.cons || [];
                
                if (pros.length === 0 && cons.length === 0) return null;
                
                return (
                  <Row gutter={16}>
                    {pros.length > 0 && (
                      <Col span={12}>
                        <Typography.Title level={5} className="text-text-primary mb-3">
                          Плюсы
                        </Typography.Title>
                        <List
                          size="small"
                          dataSource={pros}
                          renderItem={(item: string) => (
                            <List.Item className="border-b border-border-primary">
                              <Space>
                                <CheckCircleOutlined style={{ color: '#52c41a' }} />
                                <span className="text-text-primary">{item}</span>
                              </Space>
                            </List.Item>
                          )}
                        />
                      </Col>
                    )}
                    {cons.length > 0 && (
                      <Col span={pros.length > 0 ? 12 : 24}>
                        <Typography.Title level={5} className="text-text-primary mb-3">
                          Минусы
                        </Typography.Title>
                        <List
                          size="small"
                          dataSource={cons}
                          renderItem={(item: string) => (
                            <List.Item className="border-b border-border-primary">
                              <Space>
                                <CloseCircleOutlined style={{ color: '#ff4d4f' }} />
                                <span className="text-text-primary">{item}</span>
                              </Space>
                            </List.Item>
                          )}
                        />
                      </Col>
                    )}
                  </Row>
                );
              } catch {
                return <div className="text-text-secondary">{analysis.pros_cons || '—'}</div>;
              }
            })()}
          </Card>
        </>
      )}

      <Modal
        title="Редактировать фонд"
        open={editModalVisible}
        onOk={handleSaveFund}
        onCancel={() => setEditModalVisible(false)}
        okText="Сохранить"
        cancelText="Отмена"
        confirmLoading={saving}
      >
        <Form form={editForm} layout="vertical">
          <Form.Item name="name" label="Название" rules={[{ required: true, message: 'Введите название' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="isin" label="ISIN" rules={[{ required: true, message: 'Введите ISIN' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="ticker" label="Тикер">
            <Input />
          </Form.Item>
          <Form.Item name="management_company" label="Управляющая компания">
            <Input />
          </Form.Item>
          <Form.Item name="real_estate_segment" label="Сегмент недвижимости">
            <Select
              allowClear
              options={[
                { value: 'склады', label: 'Склады' },
                { value: 'офисы', label: 'Офисы' },
                { value: 'ТЦ', label: 'Торговые центры' },
                { value: 'ЦОД', label: 'ЦОД' },
                { value: 'жильё', label: 'Жильё' },
              ]}
            />
          </Form.Item>
          <Form.Item name="qualified_required" label="Требуется статус квала" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="has_market_maker" label="Маркет-мейкер" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="fund_start_date" label="Дата начала">
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="fund_end_date" label="Дата завершения">
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="investfunds_url" label="URL на investfunds.ru" tooltip="Например: https://investfunds.ru/funds/5887/">
            <Input placeholder="https://investfunds.ru/funds/..." />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default FundDetails;
