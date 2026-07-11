export function KpiCard({label,value,tone="blue"}:{label:string;value:string;tone?:"blue"|"green"|"red"}){return <div className={`kpi ${tone}`}><span>{label}</span><strong>{value}</strong></div>}
