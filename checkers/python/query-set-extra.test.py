# <expect-error>
Entry.objects.get().extra()

# <expect-error>
Entry.objects.filter().extra()

# <expect-error>
Entry.objects.update().extra()

# <expect-error>
Entry.objects.filter().update().extra()

# <expect-error>
Entry.objects.get({}).filter().update().extra()

# <no-error>
findings = Finding.objects.filter(verified=True,
                                      severity__in=('Critical', 'High', 'Medium', 'Low', 'Info')).prefetch_related(
        'test__engagement__product',
        'test__engagement__product__prod_type',
        'test__engagement__risk_acceptance',
        'risk_acceptance_set',
        'reporter').extra(
        select={
            'ra_count': 'SELECT COUNT(*) FROM dojo_risk_acceptance INNER JOIN '
                        'dojo_risk_acceptance_accepted_findings ON '
                        '( dojo_risk_acceptance.id = dojo_risk_acceptance_accepted_findings.risk_acceptance_id ) '
                        'WHERE dojo_risk_acceptance_accepted_findings.finding_id = dojo_finding.id',
        },
    )
