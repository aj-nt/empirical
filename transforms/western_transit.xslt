<?xml version="1.0" encoding="UTF-8"?>
<!--
  Western Transit XSLT

  Consumes TransitChart XML. Computes:
  - Transit-to-natal aspects (9 aspects @ 8° orb)
  - Placidus house overlays
  - 4 essential dignities (domicile, detriment, exaltation, fall)
  - Nodes and angles
-->
<xsl:stylesheet xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
                xmlns:math="http://exslt.org/math"
                xmlns:exslt="http://exslt.org/common"
                xmlns:western="urn:empirical:western"
                version="1.0"
                exclude-result-prefixes="math exslt">

  <!-- ── Parameters ────────────────────────────────────────────────────── -->
  <xsl:param name="orb" select="8.0"/>

  <!-- ── All 19 planets (classical + outer + asteroids + Lilith + Chiron + dwarfs) ── -->
  <xsl:variable name="allPlanets" select="',Sun,Moon,Mercury,Venus,Mars,Jupiter,Saturn,Uranus,Neptune,Pluto,Ceres,Pallas,Juno,Vesta,Lilith,Chiron,Eris,Makemake,Gonggong,'"/>

  <!-- ── 9 aspects ─────────────────────────────────────────────────────── -->
  <xsl:variable name="aspects">
    <aspect angle="0" name="conjunction"/>
    <aspect angle="30" name="semi-sextile"/>
    <aspect angle="45" name="semi-square"/>
    <aspect angle="60" name="sextile"/>
    <aspect angle="90" name="square"/>
    <aspect angle="120" name="trine"/>
    <aspect angle="135" name="sesquiquadrate"/>
    <aspect angle="150" name="quincunx"/>
    <aspect angle="180" name="opposition"/>
  </xsl:variable>

  <!-- ── Sign names ────────────────────────────────────────────────────── -->
  <xsl:variable name="signs" select="'Aries,Taurus,Gemini,Cancer,Leo,Virgo,Libra,Scorpio,Sagittarius,Capricorn,Aquarius,Pisces'"/>

  <!-- ── Dignity tables ────────────────────────────────────────────────── -->
  <!-- Domicile: sign index → planet name -->
  <xsl:variable name="domicile0" select="'Mars'"/>     <!-- Aries -->
  <xsl:variable name="domicile1" select="'Venus'"/>    <!-- Taurus -->
  <xsl:variable name="domicile2" select="'Mercury'"/>  <!-- Gemini -->
  <xsl:variable name="domicile3" select="'Moon'"/>     <!-- Cancer -->
  <xsl:variable name="domicile4" select="'Sun'"/>      <!-- Leo -->
  <xsl:variable name="domicile5" select="'Mercury'"/>  <!-- Virgo -->
  <xsl:variable name="domicile6" select="'Venus'"/>    <!-- Libra -->
  <xsl:variable name="domicile7" select="'Pluto'"/>    <!-- Scorpio (modern) -->
  <xsl:variable name="domicile8" select="'Jupiter'"/>  <!-- Sagittarius -->
  <xsl:variable name="domicile9" select="'Saturn'"/>   <!-- Capricorn -->
  <xsl:variable name="domicile10" select="'Uranus'"/>  <!-- Aquarius (modern) -->
  <xsl:variable name="domicile11" select="'Neptune'"/> <!-- Pisces (modern) -->

  <!-- Detriment: opposite sign of domicile -->
  <xsl:variable name="detriment0" select="'Venus'"/>   <!-- Aries → Libra's ruler -->
  <xsl:variable name="detriment1" select="'Pluto'"/>    <!-- Taurus → Scorpio's ruler -->
  <xsl:variable name="detriment2" select="'Jupiter'"/> <!-- Gemini → Sagittarius's ruler -->
  <xsl:variable name="detriment3" select="'Saturn'"/>  <!-- Cancer → Capricorn's ruler -->
  <xsl:variable name="detriment4" select="'Uranus'"/>  <!-- Leo → Aquarius's ruler -->
  <xsl:variable name="detriment5" select="'Neptune'"/> <!-- Virgo → Pisces's ruler -->
  <xsl:variable name="detriment6" select="'Mars'"/>    <!-- Libra → Aries's ruler -->
  <xsl:variable name="detriment7" select="'Venus'"/>   <!-- Scorpio → Taurus's ruler -->
  <xsl:variable name="detriment8" select="'Mercury'"/> <!-- Sagittarius → Gemini's ruler -->
  <xsl:variable name="detriment9" select="'Moon'"/>    <!-- Capricorn → Cancer's ruler -->
  <xsl:variable name="detriment10" select="'Sun'"/>    <!-- Aquarius → Leo's ruler -->
  <xsl:variable name="detriment11" select="'Mercury'"/> <!-- Pisces → Virgo's ruler -->

  <!-- Exaltation: sign index → planet name -->
  <xsl:variable name="exalt0" select="'Sun'"/>      <!-- Aries: Sun exalted -->
  <xsl:variable name="exalt1" select="'Moon'"/>     <!-- Taurus: Moon exalted -->
  <xsl:variable name="exalt2" select="''"/>         <!-- Gemini: none -->
  <xsl:variable name="exalt3" select="'Jupiter'"/>  <!-- Cancer: Jupiter exalted -->
  <xsl:variable name="exalt4" select="''"/>         <!-- Leo: none -->
  <xsl:variable name="exalt5" select="'Mercury'"/>  <!-- Virgo: Mercury exalted -->
  <xsl:variable name="exalt6" select="'Saturn'"/>   <!-- Libra: Saturn exalted -->
  <xsl:variable name="exalt7" select="''"/>         <!-- Scorpio: none -->
  <xsl:variable name="exalt8" select="''"/>         <!-- Sagittarius: none -->
  <xsl:variable name="exalt9" select="'Mars'"/>     <!-- Capricorn: Mars exalted -->
  <xsl:variable name="exalt10" select="''"/>        <!-- Aquarius: none -->
  <xsl:variable name="exalt11" select="'Venus'"/>   <!-- Pisces: Venus exalted -->

  <!-- Fall: opposite sign of exaltation -->
  <xsl:variable name="fall0" select="'Saturn'"/>    <!-- Aries: Saturn fall -->
  <xsl:variable name="fall1" select="''"/>          <!-- Taurus: none -->
  <xsl:variable name="fall2" select="''"/>          <!-- Gemini: none -->
  <xsl:variable name="fall3" select="'Mars'"/>      <!-- Cancer: Mars fall -->
  <xsl:variable name="fall4" select="''"/>          <!-- Leo: none -->
  <xsl:variable name="fall5" select="'Venus'"/>     <!-- Virgo: Venus fall -->
  <xsl:variable name="fall6" select="'Sun'"/>       <!-- Libra: Sun fall -->
  <xsl:variable name="fall7" select="'Moon'"/>      <!-- Scorpio: Moon fall -->
  <xsl:variable name="fall8" select="''"/>          <!-- Sagittarius: none -->
  <xsl:variable name="fall9" select="'Jupiter'"/>   <!-- Capricorn: Jupiter fall -->
  <xsl:variable name="fall10" select="''"/>         <!-- Aquarius: none -->
  <xsl:variable name="fall11" select="'Mercury'"/>  <!-- Pisces: Mercury fall -->

  <!-- ══════════════════════════════════════════════════════════════════════
       ROOT TEMPLATE
       ════════════════════════════════════════════════════════════════════ -->
  <xsl:template match="/TransitChart">
    <western:TransitReport xmlns:western="urn:empirical:western">
      <western:Name><xsl:value-of select="Identity/Name"/></western:Name>
      <western:TransitDate>
        <xsl:value-of select="Moment/Year"/>-<xsl:value-of select="format-number(Moment/Month,'00')"/>-<xsl:value-of select="format-number(Moment/Day,'00')"/>
      </western:TransitDate>

      <!-- Capture natal Placidus houses for overlay -->
      <xsl:variable name="natalHouses" select="Natal/Houses/System[@name='placidus']/Cusp"/>

      <!-- ══════════════════════════════════════════════════════════════════
           TRANSIT-TO-NATAL ASPECTS
           ════════════════════════════════════════════════════════════════ -->
      <western:TransitAspects>
        <xsl:for-each select="Transits/Planet[contains($allPlanets, concat(',',@name,','))]">
          <xsl:variable name="tPlanet" select="@name"/>
          <xsl:variable name="tLon" select="Tropical/Lon"/>
          <xsl:variable name="tSign" select="floor($tLon div 30)"/>

          <xsl:for-each select="/TransitChart/Natal/Positions/Planet[contains($allPlanets, concat(',',@name,','))]">
            <xsl:variable name="nPlanet" select="@name"/>
            <xsl:variable name="nLon" select="Tropical/Lon"/>
            <xsl:variable name="nSign" select="floor($nLon div 30)"/>

            <xsl:variable name="rawDist" select="$tLon - $nLon"/>
            <xsl:variable name="dist">
              <xsl:choose>
                <xsl:when test="$rawDist &lt; 0"><xsl:value-of select="$rawDist + 360"/></xsl:when>
                <xsl:otherwise><xsl:value-of select="$rawDist"/></xsl:otherwise>
              </xsl:choose>
            </xsl:variable>

            <xsl:for-each select="exslt:node-set($aspects)/aspect">
              <xsl:variable name="aspAngle" select="@angle"/>
              <xsl:variable name="aspName" select="@name"/>
              <xsl:variable name="diff">
                <xsl:choose>
                  <xsl:when test="$dist > $aspAngle">
                    <xsl:value-of select="$dist - $aspAngle"/>
                  </xsl:when>
                  <xsl:otherwise>
                    <xsl:value-of select="$aspAngle - $dist"/>
                  </xsl:otherwise>
                </xsl:choose>
              </xsl:variable>

              <xsl:if test="$diff &lt;= $orb">
                <western:TransitAspect>
                  <western:TransitPlanet><xsl:value-of select="$tPlanet"/></western:TransitPlanet>
                  <western:NatalPlanet><xsl:value-of select="$nPlanet"/></western:NatalPlanet>
                  <western:Aspect><xsl:value-of select="$aspName"/></western:Aspect>
                  <western:Orb><xsl:value-of select="format-number($diff,'0.00')"/></western:Orb>
                  <western:TransitSign>
                    <xsl:call-template name="signName">
                      <xsl:with-param name="idx" select="$tSign"/>
                    </xsl:call-template>
                  </western:TransitSign>
                  <western:NatalSign>
                    <xsl:call-template name="signName">
                      <xsl:with-param name="idx" select="$nSign"/>
                    </xsl:call-template>
                  </western:NatalSign>
                  <!-- Placidus house overlay -->
                  <western:NatalHouse>
                    <xsl:call-template name="placidusHouse">
                      <xsl:with-param name="lon" select="$tLon"/>
                      <xsl:with-param name="cusps" select="$natalHouses"/>
                    </xsl:call-template>
                  </western:NatalHouse>
                </western:TransitAspect>
              </xsl:if>
            </xsl:for-each>
          </xsl:for-each>
        </xsl:for-each>
      </western:TransitAspects>

      <!-- ══════════════════════════════════════════════════════════════════
           TRANSIT PLANET DIGNITY
           ════════════════════════════════════════════════════════════════ -->
      <western:TransitDignities>
        <xsl:for-each select="Transits/Planet[contains($allPlanets, concat(',',@name,','))]">
          <xsl:variable name="planet" select="@name"/>
          <xsl:variable name="lon" select="Tropical/Lon"/>
          <xsl:variable name="sign" select="floor($lon div 30)"/>

          <western:TransitDignity>
            <western:Planet><xsl:value-of select="$planet"/></western:Planet>
            <western:Sign>
              <xsl:call-template name="signName">
                <xsl:with-param name="idx" select="$sign"/>
              </xsl:call-template>
            </western:Sign>
            <western:Domicile>
              <xsl:call-template name="checkDignity">
                <xsl:with-param name="planet" select="$planet"/>
                <xsl:with-param name="sign" select="$sign"/>
                <xsl:with-param name="table" select="'domicile'"/>
              </xsl:call-template>
            </western:Domicile>
            <western:Detriment>
              <xsl:call-template name="checkDignity">
                <xsl:with-param name="planet" select="$planet"/>
                <xsl:with-param name="sign" select="$sign"/>
                <xsl:with-param name="table" select="'detriment'"/>
              </xsl:call-template>
            </western:Detriment>
            <western:Exaltation>
              <xsl:call-template name="checkDignity">
                <xsl:with-param name="planet" select="$planet"/>
                <xsl:with-param name="sign" select="$sign"/>
                <xsl:with-param name="table" select="'exalt'"/>
              </xsl:call-template>
            </western:Exaltation>
            <western:Fall>
              <xsl:call-template name="checkDignity">
                <xsl:with-param name="planet" select="$planet"/>
                <xsl:with-param name="sign" select="$sign"/>
                <xsl:with-param name="table" select="'fall'"/>
              </xsl:call-template>
            </western:Fall>
          </western:TransitDignity>
        </xsl:for-each>
      </western:TransitDignities>

      <!-- ══════════════════════════════════════════════════════════════════
           TRANSIT ANGLES AND NODES
           ════════════════════════════════════════════════════════════════ -->
      <western:TransitAngles>
        <western:ASC>
          <xsl:call-template name="signName">
            <xsl:with-param name="idx" select="floor(Angles/ASC div 30)"/>
          </xsl:call-template>
          <xsl:text> </xsl:text>
          <xsl:value-of select="format-number(Angles/ASC mod 30,'0.00')"/>
        </western:ASC>
        <western:MC>
          <xsl:call-template name="signName">
            <xsl:with-param name="idx" select="floor(Angles/MC div 30)"/>
          </xsl:call-template>
          <xsl:text> </xsl:text>
          <xsl:value-of select="format-number(Angles/MC mod 30,'0.00')"/>
        </western:MC>
      </western:TransitAngles>

      <western:TransitNodes>
        <western:NorthNode>
          <xsl:call-template name="signName">
            <xsl:with-param name="idx" select="floor(Nodes/NorthNode div 30)"/>
          </xsl:call-template>
          <xsl:text> </xsl:text>
          <xsl:value-of select="format-number(Nodes/NorthNode mod 30,'0.00')"/>
        </western:NorthNode>
      </western:TransitNodes>

    </western:TransitReport>
  </xsl:template>

  <!-- ══════════════════════════════════════════════════════════════════════
       HELPER TEMPLATES
       ════════════════════════════════════════════════════════════════════ -->

  <xsl:template name="signName">
    <xsl:param name="idx"/>
    <xsl:call-template name="nthToken">
      <xsl:with-param name="list" select="$signs"/>
      <xsl:with-param name="n" select="$idx + 1"/>
    </xsl:call-template>
  </xsl:template>

  <!-- Placidus house: find which house cusp range contains the longitude -->
  <xsl:template name="placidusHouse">
    <xsl:param name="lon"/>
    <xsl:param name="cusps"/>
    <xsl:variable name="c1" select="$cusps[1]"/>
    <xsl:variable name="c2" select="$cusps[2]"/>
    <xsl:variable name="c3" select="$cusps[3]"/>
    <xsl:variable name="c4" select="$cusps[4]"/>
    <xsl:variable name="c5" select="$cusps[5]"/>
    <xsl:variable name="c6" select="$cusps[6]"/>
    <xsl:variable name="c7" select="$cusps[7]"/>
    <xsl:variable name="c8" select="$cusps[8]"/>
    <xsl:variable name="c9" select="$cusps[9]"/>
    <xsl:variable name="c10" select="$cusps[10]"/>
    <xsl:variable name="c11" select="$cusps[11]"/>
    <xsl:variable name="c12" select="$cusps[12]"/>
    <xsl:choose>
      <xsl:when test="$lon >= $c1 and $lon &lt; $c2">1</xsl:when>
      <xsl:when test="$lon >= $c2 and $lon &lt; $c3">2</xsl:when>
      <xsl:when test="$lon >= $c3 and $lon &lt; $c4">3</xsl:when>
      <xsl:when test="$lon >= $c4 and $lon &lt; $c5">4</xsl:when>
      <xsl:when test="$lon >= $c5 and $lon &lt; $c6">5</xsl:when>
      <xsl:when test="$lon >= $c6 and $lon &lt; $c7">6</xsl:when>
      <xsl:when test="$lon >= $c7 and $lon &lt; $c8">7</xsl:when>
      <xsl:when test="$lon >= $c8 and $lon &lt; $c9">8</xsl:when>
      <xsl:when test="$lon >= $c9 and $lon &lt; $c10">9</xsl:when>
      <xsl:when test="$lon >= $c10 and $lon &lt; $c11">10</xsl:when>
      <xsl:when test="$lon >= $c11 and $lon &lt; $c12">11</xsl:when>
      <xsl:otherwise>12</xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <xsl:template name="checkDignity">
    <xsl:param name="planet"/>
    <xsl:param name="sign"/>
    <xsl:param name="table"/>
    <xsl:variable name="ruler">
      <xsl:choose>
        <xsl:when test="$table = 'domicile'">
          <xsl:call-template name="getDomicile">
            <xsl:with-param name="sign" select="$sign"/>
          </xsl:call-template>
        </xsl:when>
        <xsl:when test="$table = 'detriment'">
          <xsl:call-template name="getDetriment">
            <xsl:with-param name="sign" select="$sign"/>
          </xsl:call-template>
        </xsl:when>
        <xsl:when test="$table = 'exalt'">
          <xsl:call-template name="getExalt">
            <xsl:with-param name="sign" select="$sign"/>
          </xsl:call-template>
        </xsl:when>
        <xsl:otherwise>
          <xsl:call-template name="getFall">
            <xsl:with-param name="sign" select="$sign"/>
          </xsl:call-template>
        </xsl:otherwise>
      </xsl:choose>
    </xsl:variable>
    <xsl:choose>
      <xsl:when test="$planet = $ruler">true</xsl:when>
      <xsl:otherwise>false</xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <xsl:template name="getDomicile">
    <xsl:param name="sign"/>
    <xsl:choose>
      <xsl:when test="$sign = 0"><xsl:value-of select="$domicile0"/></xsl:when>
      <xsl:when test="$sign = 1"><xsl:value-of select="$domicile1"/></xsl:when>
      <xsl:when test="$sign = 2"><xsl:value-of select="$domicile2"/></xsl:when>
      <xsl:when test="$sign = 3"><xsl:value-of select="$domicile3"/></xsl:when>
      <xsl:when test="$sign = 4"><xsl:value-of select="$domicile4"/></xsl:when>
      <xsl:when test="$sign = 5"><xsl:value-of select="$domicile5"/></xsl:when>
      <xsl:when test="$sign = 6"><xsl:value-of select="$domicile6"/></xsl:when>
      <xsl:when test="$sign = 7"><xsl:value-of select="$domicile7"/></xsl:when>
      <xsl:when test="$sign = 8"><xsl:value-of select="$domicile8"/></xsl:when>
      <xsl:when test="$sign = 9"><xsl:value-of select="$domicile9"/></xsl:when>
      <xsl:when test="$sign = 10"><xsl:value-of select="$domicile10"/></xsl:when>
      <xsl:when test="$sign = 11"><xsl:value-of select="$domicile11"/></xsl:when>
    </xsl:choose>
  </xsl:template>

  <xsl:template name="getDetriment">
    <xsl:param name="sign"/>
    <xsl:choose>
      <xsl:when test="$sign = 0"><xsl:value-of select="$detriment0"/></xsl:when>
      <xsl:when test="$sign = 1"><xsl:value-of select="$detriment1"/></xsl:when>
      <xsl:when test="$sign = 2"><xsl:value-of select="$detriment2"/></xsl:when>
      <xsl:when test="$sign = 3"><xsl:value-of select="$detriment3"/></xsl:when>
      <xsl:when test="$sign = 4"><xsl:value-of select="$detriment4"/></xsl:when>
      <xsl:when test="$sign = 5"><xsl:value-of select="$detriment5"/></xsl:when>
      <xsl:when test="$sign = 6"><xsl:value-of select="$detriment6"/></xsl:when>
      <xsl:when test="$sign = 7"><xsl:value-of select="$detriment7"/></xsl:when>
      <xsl:when test="$sign = 8"><xsl:value-of select="$detriment8"/></xsl:when>
      <xsl:when test="$sign = 9"><xsl:value-of select="$detriment9"/></xsl:when>
      <xsl:when test="$sign = 10"><xsl:value-of select="$detriment10"/></xsl:when>
      <xsl:when test="$sign = 11"><xsl:value-of select="$detriment11"/></xsl:when>
    </xsl:choose>
  </xsl:template>

  <xsl:template name="getExalt">
    <xsl:param name="sign"/>
    <xsl:choose>
      <xsl:when test="$sign = 0"><xsl:value-of select="$exalt0"/></xsl:when>
      <xsl:when test="$sign = 1"><xsl:value-of select="$exalt1"/></xsl:when>
      <xsl:when test="$sign = 2"><xsl:value-of select="$exalt2"/></xsl:when>
      <xsl:when test="$sign = 3"><xsl:value-of select="$exalt3"/></xsl:when>
      <xsl:when test="$sign = 4"><xsl:value-of select="$exalt4"/></xsl:when>
      <xsl:when test="$sign = 5"><xsl:value-of select="$exalt5"/></xsl:when>
      <xsl:when test="$sign = 6"><xsl:value-of select="$exalt6"/></xsl:when>
      <xsl:when test="$sign = 7"><xsl:value-of select="$exalt7"/></xsl:when>
      <xsl:when test="$sign = 8"><xsl:value-of select="$exalt8"/></xsl:when>
      <xsl:when test="$sign = 9"><xsl:value-of select="$exalt9"/></xsl:when>
      <xsl:when test="$sign = 10"><xsl:value-of select="$exalt10"/></xsl:when>
      <xsl:when test="$sign = 11"><xsl:value-of select="$exalt11"/></xsl:when>
    </xsl:choose>
  </xsl:template>

  <xsl:template name="getFall">
    <xsl:param name="sign"/>
    <xsl:choose>
      <xsl:when test="$sign = 0"><xsl:value-of select="$fall0"/></xsl:when>
      <xsl:when test="$sign = 1"><xsl:value-of select="$fall1"/></xsl:when>
      <xsl:when test="$sign = 2"><xsl:value-of select="$fall2"/></xsl:when>
      <xsl:when test="$sign = 3"><xsl:value-of select="$fall3"/></xsl:when>
      <xsl:when test="$sign = 4"><xsl:value-of select="$fall4"/></xsl:when>
      <xsl:when test="$sign = 5"><xsl:value-of select="$fall5"/></xsl:when>
      <xsl:when test="$sign = 6"><xsl:value-of select="$fall6"/></xsl:when>
      <xsl:when test="$sign = 7"><xsl:value-of select="$fall7"/></xsl:when>
      <xsl:when test="$sign = 8"><xsl:value-of select="$fall8"/></xsl:when>
      <xsl:when test="$sign = 9"><xsl:value-of select="$fall9"/></xsl:when>
      <xsl:when test="$sign = 10"><xsl:value-of select="$fall10"/></xsl:when>
      <xsl:when test="$sign = 11"><xsl:value-of select="$fall11"/></xsl:when>
    </xsl:choose>
  </xsl:template>

  <xsl:template name="nthToken">
    <xsl:param name="list"/>
    <xsl:param name="n"/>
    <xsl:param name="delimiter" select="','"/>
    <xsl:choose>
      <xsl:when test="$n = 1">
        <xsl:choose>
          <xsl:when test="contains($list, $delimiter)">
            <xsl:value-of select="substring-before($list, $delimiter)"/>
          </xsl:when>
          <xsl:otherwise>
            <xsl:value-of select="$list"/>
          </xsl:otherwise>
        </xsl:choose>
      </xsl:when>
      <xsl:when test="contains($list, $delimiter)">
        <xsl:call-template name="nthToken">
          <xsl:with-param name="list" select="substring-after($list, $delimiter)"/>
          <xsl:with-param name="n" select="$n - 1"/>
          <xsl:with-param name="delimiter" select="$delimiter"/>
        </xsl:call-template>
      </xsl:when>
    </xsl:choose>
  </xsl:template>

</xsl:stylesheet>
