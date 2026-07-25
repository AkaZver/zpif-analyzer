import React, { useCallback, useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Typography, Card, Row, Col, Statistic, Tag, Button, Space, Table,
  message, Spin, Upload, Descriptions, List, Modal, Form, Input, Select,
  Switch, Popconfirm, DatePicker, Segmented, Tooltip,
} from 'antd';
import { ArrowLeftOutlined, SearchOutlined, UploadOutlined, ThunderboltOutlined, CheckCircleOutlined, CloseCircleOutlined, DeleteOutlined, DownloadOutlined, EditOutlined, CloudDownloadOutlined, QuestionCircleOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import {
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip as ChartTooltip, Legend, ResponsiveContainer,
  BarChart, Bar, ReferenceLine,
} from 'recharts';
import { apiClient } from '../../api/client';
import type { Fund, FundFinancials, FundDocument, LLMAnalysis } from '../../types';
import { formatMonthYear } from '../../utils/dateFormatters';
import {
  buildPayoutChartData,
  getTradingStartFormatted,
  groupFinancialsByMonth,
} from '../../utils/chartDataTransformers';
import { useTheme } from '../../hooks/useTheme';

const FundDetails: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { theme } = useTheme();
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
  const [timeRange, setTimeRange] = useState<'3m' | '6m' | '1y' | '3y' | '5y' | 'all'>('1y');
  const [selectedDocIds, setSelectedDocIds] = useState<number[]>([]);

  const formatNumber = (value: number | string, digits = 0) =>
    new Intl.NumberFormat('ru-RU', {
      minimumFractionDigits: digits,
      maximumFractionDigits: digits,
    }).format(Number(value));

  const loadDocuments = useCallback(async () => {
    if (!id) return;
    try {
      const documentsData = await apiClient.getDocuments(parseInt(id));
      setDocuments(documentsData);
      setSelectedDocIds(documentsData.filter(d => d.status !== 'analyzed').map(d => d.id));
    } catch {
      message.error('Не удалось загрузить документы');
    }
  }, [id]);

  const loadData = useCallback(async () => {
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
      setSelectedDocIds(documentsData.filter(d => d.status !== 'analyzed').map(d => d.id));
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
  }, [id]);

  useEffect(() => {
    if (id) loadData();
  }, [id, loadData]);

  const handleDiscover = async () => {
    if (!id) return;
    setDiscovering(true);
    try {
      await apiClient.discoverDocuments(parseInt(id));
      message.success('Автопоиск документов запущен');
      setTimeout(() => loadDocuments(), 3000);
    } catch (error: any) {
      const errorMsg = error?.response?.data?.error || 'Ошибка при поиске документов';
      message.error(errorMsg, 8);
    } finally {
      setDiscovering(false);
    }
  };

  const handleUpload = async (file: File) => {
    if (!id) return false;
    try {
      await apiClient.uploadDocument(parseInt(id), file);
      message.success('Документ загружен');
      await loadDocuments();
    } catch {
      message.error('Ошибка при загрузке');
    }
    return false;
  };

  const handleAnalyze = async () => {
    if (!id) return;
    setAnalyzing(true);
    try {
      const result = await apiClient.analyzeFund(parseInt(id), selectedDocIds);
      setAnalysis(result);
      message.success('Анализ завершён');
      await loadData();
    } catch (error: any) {
      const errorMsg = error?.response?.data?.error || 'Ошибка при анализе';
      message.error(errorMsg, 8);
    } finally {
      setAnalyzing(false);
    }
  };

  const handleDeleteDocument = async (docId: number) => {
    if (!id) return;
    try {
      await apiClient.deleteDocument(parseInt(id), docId);
      message.success('Документ удалён');
      await loadDocuments();
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

  const filteredFinancials = (() => {
    if (timeRange === 'all') return financials;
    const months: Record<string, number> = { '3m': 3, '6m': 6, '1y': 12, '3y': 36, '5y': 60 };
    const cutoff = dayjs().subtract(months[timeRange], 'month');
    return financials.filter(f => dayjs(f.snapshot_date).isAfter(cutoff));
  })();

  const tradingStartFormatted = getTradingStartFormatted(financials);

  const priceChartData = (() => {
    const grouped = groupFinancialsByMonth(
      filteredFinancials,
      (f) => f.unit_price_rub > 0 || f.nav_per_unit_rub > 0
    );

    return grouped.map((f) => {
      const currentDate = new Date(f.snapshot_date);
      const formattedDate = formatMonthYear(currentDate);
      const showPrice = tradingStartDate && currentDate >= tradingStartDate;

      return {
        date: formattedDate,
        'Цена пая': showPrice ? f.unit_price_rub : null,
        'РСП': f.nav_per_unit_rub,
      };
    });
  })();

  const payoutChartData = buildPayoutChartData(filteredFinancials);

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

  const renderMetricTitle = (title: string, description: string) => (
    <Tooltip title={description}>
      <span>{title} <QuestionCircleOutlined style={{ color: '#888' }} /></span>
    </Tooltip>
  );

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
        <Descriptions.Item label="Дата завершения">
          {fund.fund_end_date ? new Date(fund.fund_end_date).toLocaleDateString('ru-RU') : '—'}
        </Descriptions.Item>
      </Descriptions>

      <Typography.Title level={4} className="text-text-primary mb-4">
        Ключевые метрики
      </Typography.Title>
      <Row gutter={[16, 16]} className="mb-6">
        <Col xs={12} sm={8} md={6}>
          <Card className="bg-white dark:bg-[#333333] border-0">
            <Statistic title={renderMetricTitle("Цена пая", "Рыночная цена одного пая на бирже")} value={latest?.unit_price_rub || 0} suffix="₽" formatter={(value: any) => formatNumber(value, 0)} />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card className="bg-white dark:bg-[#333333] border-0">
            <Statistic title={renderMetricTitle("РСП", "Расчётная стоимость пая (NAV на пай)")} value={latest?.nav_per_unit_rub || 0} suffix="₽" formatter={(value: any) => formatNumber(value, 0)} />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card className="bg-white dark:bg-[#333333] border-0">
            <Statistic
              title={renderMetricTitle("Дисконт к РСП", "Разница между рыночной ценой и РСП в процентах")}
              value={latest?.discount_to_nav_pct || 0}
              suffix="%"
              formatter={(value: any) => formatNumber(value, 1)}
              valueStyle={{ color: (latest?.discount_to_nav_pct || 0) <= 0 ? '#52c41a' : '#ff4d4f' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card className="bg-white dark:bg-[#333333] border-0">
            <Statistic title={renderMetricTitle("Cap Rate", "Коэффициент капитализации (NOI / Стоимость активов)")} value={latest?.cap_rate_pct || 0} suffix="%" formatter={(value: any) => formatNumber(value, 1)} />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card className="bg-white dark:bg-[#333333] border-0">
            <Statistic 
              title={renderMetricTitle("СЧА", "Стоимость чистых активов фонда")}
              value={latest?.nav_total_mln_rub || 0} 
              suffix="млн ₽" 
              formatter={(value: any) => formatNumber(value, 2)}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card className="bg-white dark:bg-[#333333] border-0">
            <Statistic title={renderMetricTitle("P/NAV", "Отношение рыночной цены к NAV")} value={latest?.p_nav || 0} formatter={(value: any) => formatNumber(value, 2)} />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card className="bg-white dark:bg-[#333333] border-0">
            <Statistic title={renderMetricTitle("P/AFFO", "Отношение рыночной цены к AFFO")} value={latest?.p_affo || 0} formatter={(value: any) => formatNumber(value, 2)} />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card className="bg-white dark:bg-[#333333] border-0">
            <Statistic
              title={renderMetricTitle("Доходность выплат", "Годовая доходность от выплат дивидендов")}
              value={latest?.payout_yield_pct || 0}
              suffix="%"
              formatter={(value: any) => formatNumber(value, 1)}
              valueStyle={{ color: (latest?.payout_yield_pct || 0) >= 0 ? '#52c41a' : '#ff4d4f' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card className="bg-white dark:bg-[#333333] border-0">
            <Statistic title={renderMetricTitle("Комиссия УК", "Ежегодная комиссия управляющей компании")} value={latest?.management_fee_pct || 0} suffix="%" formatter={(value: any) => formatNumber(value, 1)} />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card className="bg-white dark:bg-[#333333] border-0">
            <Statistic title={renderMetricTitle("Объектов", "Количество объектов недвижимости в портфеле")} value={latest?.number_of_properties || 0} formatter={(value: any) => formatNumber(value, 0)} />
          </Card>
        </Col>
      </Row>

      <div className="flex justify-end mb-4">
        <Segmented
          value={timeRange}
          onChange={(value) => setTimeRange(value as typeof timeRange)}
          options={[
            { label: '3 мес', value: '3m' },
            { label: '6 мес', value: '6m' },
            { label: '1 год', value: '1y' },
            { label: '3 года', value: '3y' },
            { label: '5 лет', value: '5y' },
            { label: 'Всё', value: 'all' },
          ]}
        />
      </div>

      {priceChartData.length > 0 && (
        <>
          <Typography.Title level={4} className="text-text-primary mb-4">
            Динамика цены и РСП
          </Typography.Title>
          <Card className="bg-white dark:bg-[#333333] border-0 mb-6">
            <ResponsiveContainer width="100%" height={350}>
              <LineChart data={priceChartData} margin={{ bottom: 30 }}>
                <CartesianGrid strokeDasharray="3 3" stroke={theme === 'dark' ? '#444444' : '#e0e0e0'} />
                <XAxis dataKey="date" stroke={theme === 'dark' ? '#a0a0a0' : '#666666'} interval={0} angle={-45} textAnchor="end" dy={10} />
                <YAxis stroke={theme === 'dark' ? '#a0a0a0' : '#666666'} tickFormatter={(value) => new Intl.NumberFormat('ru-RU').format(value)} />
                <ChartTooltip formatter={(value: any) => formatNumber(value, 2)} contentStyle={{ backgroundColor: theme === 'dark' ? '#333333' : '#ffffff', border: 'none' }} />
                <Legend layout="horizontal" verticalAlign="top" align="center" />
                
                {tradingStartFormatted && priceChartData.some(d => d.date === tradingStartFormatted) && (
                  <ReferenceLine 
                    x={tradingStartFormatted} 
                    stroke={theme === 'dark' ? '#888888' : '#999999'} 
                    strokeDasharray="3 3"
                    label={{ 
                      value: 'Начало торгов', 
                      position: priceChartData.findIndex(d => d.date === tradingStartFormatted) < priceChartData.length / 2 ? 'right' : 'left',
                      fill: theme === 'dark' ? '#a0a0a0' : '#666666',
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
          <Card className="bg-white dark:bg-[#333333] border-0 mb-6">
            <ResponsiveContainer width="100%" height={250}>
              <BarChart data={payoutChartData}>
                <CartesianGrid strokeDasharray="3 3" stroke={theme === 'dark' ? '#444444' : '#e0e0e0'} />
                <XAxis dataKey="date" stroke={theme === 'dark' ? '#a0a0a0' : '#666666'} />
                <YAxis stroke={theme === 'dark' ? '#a0a0a0' : '#666666'} tickFormatter={(value) => new Intl.NumberFormat('ru-RU').format(value)} />
                <ChartTooltip 
                  formatter={(value: any) => formatNumber(value, 2)}
                  contentStyle={{ backgroundColor: theme === 'dark' ? '#333333' : '#ffffff', border: 'none' }}
                  cursor={{ fill: theme === 'dark' ? '#444444' : '#f0f0f0', fillOpacity: 0.3 }}
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
      <Card className="bg-white dark:bg-[#333333] border-0 mb-6">
        <Space className="mb-4">
          <Button icon={<SearchOutlined />} onClick={handleDiscover} loading={discovering}>
            Найти в интернете
          </Button>
          <Upload beforeUpload={handleUpload} showUploadList={false} accept=".pdf,.doc,.docx,.xlsx,.txt">
            <Button icon={<UploadOutlined />}>Загрузить вручную</Button>
          </Upload>
          <Button
            type="primary"
            icon={<ThunderboltOutlined />}
            onClick={handleAnalyze}
            loading={analyzing}
            disabled={selectedDocIds.length === 0}
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
          rowSelection={{
            selectedRowKeys: selectedDocIds,
            onChange: (keys) => setSelectedDocIds(keys as number[]),
          }}
          onRow={(record) => ({
            onClick: () => {
              setSelectedDocIds(prev =>
                prev.includes(record.id)
                  ? prev.filter(id => id !== record.id)
                  : [...prev, record.id]
              );
            },
            style: { cursor: 'pointer' },
          })}
        />
      </Card>

      {analysis && (
        <>
          <Typography.Title level={4} className="text-text-primary mb-4">
            LLM-анализ
          </Typography.Title>
          <Card className="bg-white dark:bg-[#333333] border-0">
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
          <Form.Item name="fund_end_date" label="Дата завершения">
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="investfunds_url" label="URL на investfunds.ru" tooltip="Например: https://investfunds.ru/funds/5887/">
            <Input placeholder="https://investfunds.ru/funds/..." />
          </Form.Item>
          <Form.Item name="vsezpif_url" label="URL на vsezpif.ru" tooltip="Например: https://vsezpif.ru/?route=fund&id=1">
            <Input placeholder="https://vsezpif.ru/?route=fund&id=..." />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default FundDetails;
